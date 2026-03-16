package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
	"github.com/nixonmkindi/go-clean-auth-api/internal/repository"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (owner_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, query, project.OwnerID, project.Name, project.Description).
		Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *ProjectRepository) GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Project, error) {
	query := `
		SELECT id, owner_id, name, description, created_at, updated_at, deleted_at
		FROM projects
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var project domain.Project
	err := r.pool.QueryRow(ctx, query, id, ownerID).Scan(
		&project.ID,
		&project.OwnerID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) ListByOwner(ctx context.Context, params repository.ProjectListParams) ([]domain.Project, int64, error) {
	allowed := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"name":       "name",
	}
	sortColumn, sortOrder := sanitizeSort(params.SortBy, params.SortOrder, allowed, "created_at")
	offset := (params.Page - 1) * params.Limit

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	countQuery := `
		SELECT COUNT(*)
		FROM projects
		WHERE owner_id = $1 AND deleted_at IS NULL`

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, params.OwnerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, owner_id, name, description, created_at, updated_at, deleted_at
		FROM projects
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, sortColumn, sortOrder)

	rows, err := r.pool.Query(ctx, query, params.OwnerID, params.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Project, 0)
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $1,
			description = $2,
			updated_at = NOW()
		WHERE id = $3 AND owner_id = $4 AND deleted_at IS NULL
		RETURNING updated_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, query, project.Name, project.Description, project.ID, project.OwnerID).
		Scan(&project.UpdatedAt)
}

func (r *ProjectRepository) SoftDelete(ctx context.Context, id, ownerID int64) error {
	query := `
		UPDATE projects
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd, err := r.pool.Exec(ctx, query, id, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}
