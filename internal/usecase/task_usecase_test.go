package usecase

import (
	"context"
	"testing"

	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
	"github.com/nixonmkindi/go-clean-auth-api/internal/pkg/apperror"
	"github.com/nixonmkindi/go-clean-auth-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTaskRepo struct {
	listForOwnerFn func(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error)
}

func (m *mockTaskRepo) Create(ctx context.Context, task *domain.Task) error { return nil }
func (m *mockTaskRepo) GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Task, error) {
	return nil, nil
}
func (m *mockTaskRepo) ListForOwner(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error) {
	if m.listForOwnerFn != nil {
		return m.listForOwnerFn(ctx, params)
	}
	return nil, 0, nil
}
func (m *mockTaskRepo) ListByProjectForOwner(ctx context.Context, projectID, ownerID int64, page, limit int, sortBy, sortOrder string) ([]domain.Task, int64, error) {
	return nil, 0, nil
}
func (m *mockTaskRepo) Update(ctx context.Context, task *domain.Task) error     { return nil }
func (m *mockTaskRepo) SoftDelete(ctx context.Context, id, ownerID int64) error { return nil }

type mockProjectRepo struct{}

func (m *mockProjectRepo) Create(ctx context.Context, project *domain.Project) error { return nil }
func (m *mockProjectRepo) GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Project, error) {
	return &domain.Project{ID: id, OwnerID: ownerID}, nil
}
func (m *mockProjectRepo) ListByOwner(ctx context.Context, params repository.ProjectListParams) ([]domain.Project, int64, error) {
	return nil, 0, nil
}
func (m *mockProjectRepo) Update(ctx context.Context, project *domain.Project) error { return nil }
func (m *mockProjectRepo) SoftDelete(ctx context.Context, id, ownerID int64) error   { return nil }

type mockUsersRepo struct{}

func (m *mockUsersRepo) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUsersRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUsersRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return &domain.User{ID: id}, nil
}

func TestTaskUsecaseListPagination(t *testing.T) {
	captured := domain.TaskListParams{}
	tasks := &mockTaskRepo{
		listForOwnerFn: func(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error) {
			captured = params
			return []domain.Task{{ID: 1, Title: "T1", Status: domain.TaskStatusPending}}, 101, nil
		},
	}
	u := NewTaskUsecase(tasks, &mockProjectRepo{}, &mockUsersRepo{}, 10, 50)

	items, meta, err := u.List(context.Background(), ListTasksInput{
		OwnerID:   5,
		Page:      0,
		Limit:     1000,
		SortBy:    "created_at",
		SortOrder: "desc",
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 1, captured.Page)
	assert.Equal(t, 50, captured.Limit)
	assert.Equal(t, int64(101), meta.Total)
	assert.Equal(t, 3, meta.TotalPages)
}

func TestTaskUsecaseInvalidStatus(t *testing.T) {
	tasks := &mockTaskRepo{}
	u := NewTaskUsecase(tasks, &mockProjectRepo{}, &mockUsersRepo{}, 10, 50)

	status := domain.TaskStatus("blocked")
	_, _, err := u.List(context.Background(), ListTasksInput{OwnerID: 1, Status: &status})
	require.Error(t, err)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_TASK_STATUS", appErr.Code)
}
