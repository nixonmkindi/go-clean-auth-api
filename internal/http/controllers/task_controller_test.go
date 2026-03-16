package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
	custommw "github.com/nixonmkindi/go-clean-auth-api/internal/http/middleware"
	customvalidator "github.com/nixonmkindi/go-clean-auth-api/internal/pkg/validator"
	"github.com/nixonmkindi/go-clean-auth-api/internal/repository"
	"github.com/nixonmkindi/go-clean-auth-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type taskRepoMock struct {
	listForOwnerFn func(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error)
}

func (m *taskRepoMock) Create(ctx context.Context, task *domain.Task) error { return nil }
func (m *taskRepoMock) GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Task, error) {
	return nil, pgx.ErrNoRows
}
func (m *taskRepoMock) ListForOwner(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error) {
	if m.listForOwnerFn != nil {
		return m.listForOwnerFn(ctx, params)
	}
	return nil, 0, nil
}
func (m *taskRepoMock) ListByProjectForOwner(ctx context.Context, projectID, ownerID int64, page, limit int, sortBy, sortOrder string) ([]domain.Task, int64, error) {
	return nil, 0, nil
}
func (m *taskRepoMock) Update(ctx context.Context, task *domain.Task) error     { return nil }
func (m *taskRepoMock) SoftDelete(ctx context.Context, id, ownerID int64) error { return nil }

type projectRepoMock struct{}

func (m *projectRepoMock) Create(ctx context.Context, project *domain.Project) error { return nil }
func (m *projectRepoMock) GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Project, error) {
	return &domain.Project{ID: id, OwnerID: ownerID}, nil
}
func (m *projectRepoMock) ListByOwner(ctx context.Context, params repository.ProjectListParams) ([]domain.Project, int64, error) {
	return nil, 0, nil
}
func (m *projectRepoMock) Update(ctx context.Context, project *domain.Project) error { return nil }
func (m *projectRepoMock) SoftDelete(ctx context.Context, id, ownerID int64) error   { return nil }

type usersRepoMock struct{}

func (m *usersRepoMock) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *usersRepoMock) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, pgx.ErrNoRows
}
func (m *usersRepoMock) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return &domain.User{ID: id}, nil
}

func TestTaskControllerList(t *testing.T) {
	tasks := &taskRepoMock{
		listForOwnerFn: func(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error) {
			return []domain.Task{{ID: 99, Title: "Investigate", Status: domain.TaskStatusPending}}, 1, nil
		},
	}
	u := usecase.NewTaskUsecase(tasks, &projectRepoMock{}, &usersRepoMock{}, 10, 50)
	controller := NewTaskController(u)

	e := echo.New()
	e.Validator = customvalidator.New()
	e.GET("/tasks", controller.List, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(custommw.ContextUserIDKey, int64(7))
			return next(c)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks?page=1&limit=10&status=pending", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, true, body["success"])
	assert.NotNil(t, body["meta"])
}
