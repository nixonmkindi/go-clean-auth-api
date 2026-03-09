package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yourusername/go-clean-auth-api/internal/domain"
	"github.com/yourusername/go-clean-auth-api/internal/pkg/apperror"
	"github.com/yourusername/go-clean-auth-api/internal/repository"
)

type TaskUsecase struct {
	tasks        repository.TaskRepository
	projects     repository.ProjectRepository
	users        repository.UserRepository
	defaultLimit int
	maxLimit     int
}

type CreateTaskInput struct {
	ProjectID   int64
	OwnerID     int64
	AssigneeID  *int64
	Title       string
	Description string
	Status      domain.TaskStatus
	DueDate     *time.Time
}

type UpdateTaskInput struct {
	TaskID      int64
	OwnerID     int64
	AssigneeID  *int64
	Title       string
	Description string
	Status      domain.TaskStatus
	DueDate     *time.Time
}

type ListTasksInput struct {
	OwnerID   int64
	ProjectID *int64
	Assignee  *int64
	Status    *domain.TaskStatus
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

func NewTaskUsecase(tasks repository.TaskRepository, projects repository.ProjectRepository, users repository.UserRepository, defaultLimit, maxLimit int) *TaskUsecase {
	return &TaskUsecase{
		tasks:        tasks,
		projects:     projects,
		users:        users,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
	}
}

func (u *TaskUsecase) Create(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
	if _, err := u.projects.GetByIDForOwner(ctx, input.ProjectID, input.OwnerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(404, "PROJECT_NOT_FOUND", "project not found", nil, err)
		}
		return nil, apperror.New(500, "PROJECT_FETCH_ERROR", "failed to verify project", nil, err)
	}
	if err := validateStatus(input.Status); err != nil {
		return nil, err
	}
	if input.AssigneeID != nil {
		if _, err := u.users.GetByID(ctx, *input.AssigneeID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, apperror.New(400, "ASSIGNEE_NOT_FOUND", "assignee user does not exist", nil, err)
			}
			return nil, apperror.New(500, "ASSIGNEE_LOOKUP_ERROR", "failed to lookup assignee", nil, err)
		}
	}

	task := &domain.Task{
		ProjectID:   input.ProjectID,
		AssigneeID:  input.AssigneeID,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Status:      input.Status,
		DueDate:     input.DueDate,
	}

	if err := u.tasks.Create(ctx, task); err != nil {
		return nil, apperror.New(500, "TASK_CREATE_ERROR", "failed to create task", nil, err)
	}

	return task, nil
}

func (u *TaskUsecase) Get(ctx context.Context, taskID, ownerID int64) (*domain.Task, error) {
	task, err := u.tasks.GetByIDForOwner(ctx, taskID, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(404, "TASK_NOT_FOUND", "task not found", nil, err)
		}
		return nil, apperror.New(500, "TASK_FETCH_ERROR", "failed to fetch task", nil, err)
	}
	return task, nil
}

func (u *TaskUsecase) List(ctx context.Context, input ListTasksInput) ([]domain.Task, PaginationMeta, error) {
	page, limit := NormalizePagination(input.Page, input.Limit, u.defaultLimit, u.maxLimit)
	if input.Status != nil {
		if err := validateStatus(*input.Status); err != nil {
			return nil, PaginationMeta{}, err
		}
	}

	items, total, err := u.tasks.ListForOwner(ctx, domain.TaskListParams{
		OwnerID:   input.OwnerID,
		ProjectID: input.ProjectID,
		Assignee:  input.Assignee,
		Status:    input.Status,
		SortBy:    input.SortBy,
		SortOrder: input.SortOrder,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		return nil, PaginationMeta{}, apperror.New(500, "TASK_LIST_ERROR", "failed to list tasks", nil, err)
	}

	return items, BuildPaginationMeta(page, limit, total), nil
}

func (u *TaskUsecase) Update(ctx context.Context, input UpdateTaskInput) (*domain.Task, error) {
	if err := validateStatus(input.Status); err != nil {
		return nil, err
	}
	if input.AssigneeID != nil {
		if _, err := u.users.GetByID(ctx, *input.AssigneeID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, apperror.New(400, "ASSIGNEE_NOT_FOUND", "assignee user does not exist", nil, err)
			}
			return nil, apperror.New(500, "ASSIGNEE_LOOKUP_ERROR", "failed to lookup assignee", nil, err)
		}
	}

	task, err := u.Get(ctx, input.TaskID, input.OwnerID)
	if err != nil {
		return nil, err
	}

	task.AssigneeID = input.AssigneeID
	task.Title = strings.TrimSpace(input.Title)
	task.Description = strings.TrimSpace(input.Description)
	task.Status = input.Status
	task.DueDate = input.DueDate

	if err := u.tasks.Update(ctx, task); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(404, "TASK_NOT_FOUND", "task not found", nil, err)
		}
		return nil, apperror.New(500, "TASK_UPDATE_ERROR", "failed to update task", nil, err)
	}

	return task, nil
}

func (u *TaskUsecase) Assign(ctx context.Context, taskID, ownerID, assigneeID int64) (*domain.Task, error) {
	task, err := u.Get(ctx, taskID, ownerID)
	if err != nil {
		return nil, err
	}
	if _, err := u.users.GetByID(ctx, assigneeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(400, "ASSIGNEE_NOT_FOUND", "assignee user does not exist", nil, err)
		}
		return nil, apperror.New(500, "ASSIGNEE_LOOKUP_ERROR", "failed to lookup assignee", nil, err)
	}

	task.AssigneeID = &assigneeID
	if err := u.tasks.Update(ctx, task); err != nil {
		return nil, apperror.New(500, "TASK_ASSIGN_ERROR", "failed to assign task", nil, err)
	}

	return task, nil
}

func (u *TaskUsecase) MarkStatus(ctx context.Context, taskID, ownerID int64, status domain.TaskStatus) (*domain.Task, error) {
	if err := validateStatus(status); err != nil {
		return nil, err
	}

	task, err := u.Get(ctx, taskID, ownerID)
	if err != nil {
		return nil, err
	}
	task.Status = status

	if err := u.tasks.Update(ctx, task); err != nil {
		return nil, apperror.New(500, "TASK_STATUS_UPDATE_ERROR", "failed to update task status", nil, err)
	}

	return task, nil
}

func (u *TaskUsecase) Delete(ctx context.Context, taskID, ownerID int64) error {
	if err := u.tasks.SoftDelete(ctx, taskID, ownerID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return apperror.New(404, "TASK_NOT_FOUND", "task not found", nil, err)
		}
		return apperror.New(500, "TASK_DELETE_ERROR", "failed to delete task", nil, err)
	}
	return nil
}

func (u *TaskUsecase) ListForAPIClient(ctx context.Context, projectID, ownerID int64, page, limit int, sortBy, sortOrder string) ([]domain.Task, PaginationMeta, error) {
	page, limit = NormalizePagination(page, limit, u.defaultLimit, u.maxLimit)

	if _, err := u.projects.GetByIDForOwner(ctx, projectID, ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, PaginationMeta{}, apperror.New(404, "PROJECT_NOT_FOUND", "project not found", nil, err)
		}
		return nil, PaginationMeta{}, apperror.New(500, "PROJECT_FETCH_ERROR", "failed to verify project", nil, err)
	}

	items, total, err := u.tasks.ListByProjectForOwner(ctx, projectID, ownerID, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, PaginationMeta{}, apperror.New(500, "TASK_LIST_ERROR", "failed to list tasks", nil, err)
	}

	return items, BuildPaginationMeta(page, limit, total), nil
}

func validateStatus(status domain.TaskStatus) error {
	switch status {
	case domain.TaskStatusPending, domain.TaskStatusInProgress, domain.TaskStatusDone:
		return nil
	default:
		return apperror.New(400, "INVALID_TASK_STATUS", "invalid task status", map[string]string{
			"allowed": "pending,in_progress,done",
		}, nil)
	}
}
