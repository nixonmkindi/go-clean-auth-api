package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (project_id, assignee_id, title, description, status, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, query, task.ProjectID, task.AssigneeID, task.Title, task.Description, task.Status, task.DueDate).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *TaskRepository) GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Task, error) {
	query := `
		SELECT t.id, t.project_id, t.assignee_id, t.title, t.description, t.status, t.due_date, t.created_at, t.updated_at, t.deleted_at
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND p.owner_id = $2 AND p.deleted_at IS NULL AND t.deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var task domain.Task
	err := r.pool.QueryRow(ctx, query, id, ownerID).Scan(
		&task.ID,
		&task.ProjectID,
		&task.AssigneeID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepository) ListForOwner(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error) {
	allowed := map[string]string{
		"created_at": "t.created_at",
		"updated_at": "t.updated_at",
		"title":      "t.title",
		"status":     "t.status",
		"due_date":   "t.due_date",
	}
	sortColumn, sortOrder := sanitizeSort(params.SortBy, params.SortOrder, allowed, "t.created_at")
	offset := (params.Page - 1) * params.Limit

	filters := []string{"p.owner_id = $1", "p.deleted_at IS NULL", "t.deleted_at IS NULL"}
	args := []interface{}{params.OwnerID}
	argIdx := 2

	if params.ProjectID != nil {
		filters = append(filters, fmt.Sprintf("t.project_id = $%d", argIdx))
		args = append(args, *params.ProjectID)
		argIdx++
	}
	if params.Assignee != nil {
		filters = append(filters, fmt.Sprintf("t.assignee_id = $%d", argIdx))
		args = append(args, *params.Assignee)
		argIdx++
	}
	if params.Status != nil {
		filters = append(filters, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, *params.Status)
		argIdx++
	}

	whereClause := strings.Join(filters, " AND ")

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE %s`, whereClause)

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.project_id, t.assignee_id, t.title, t.description, t.status, t.due_date, t.created_at, t.updated_at, t.deleted_at
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`, whereClause, sortColumn, sortOrder, argIdx, argIdx+1)

	args = append(args, params.Limit, offset)
	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Task, 0)
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, task)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *TaskRepository) ListByProjectForOwner(ctx context.Context, projectID, ownerID int64, page, limit int, sortBy, sortOrder string) ([]domain.Task, int64, error) {
	return r.ListForOwner(ctx, domain.TaskListParams{
		OwnerID:   ownerID,
		ProjectID: &projectID,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		Limit:     limit,
	})
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET assignee_id = $1,
			title = $2,
			description = $3,
			status = $4,
			due_date = $5,
			updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING updated_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, query,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		task.ID,
	).Scan(&task.UpdatedAt)
}

func (r *TaskRepository) SoftDelete(ctx context.Context, id, ownerID int64) error {
	query := `
		UPDATE tasks t
		SET deleted_at = NOW(), updated_at = NOW()
		FROM projects p
		WHERE t.id = $1 AND p.id = t.project_id AND p.owner_id = $2 AND t.deleted_at IS NULL AND p.deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd, err := r.pool.Exec(ctx, query, id, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
