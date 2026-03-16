package repository

import (
	"context"

	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type APIClientRepository interface {
	Create(ctx context.Context, client *domain.APIClient) error
	GetByKey(ctx context.Context, key string) (*domain.APIClient, error)
}

type ProjectListParams struct {
	OwnerID   int64
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Project, error)
	ListByOwner(ctx context.Context, params ProjectListParams) ([]domain.Project, int64, error)
	Update(ctx context.Context, project *domain.Project) error
	SoftDelete(ctx context.Context, id, ownerID int64) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByIDForOwner(ctx context.Context, id, ownerID int64) (*domain.Task, error)
	ListForOwner(ctx context.Context, params domain.TaskListParams) ([]domain.Task, int64, error)
	ListByProjectForOwner(ctx context.Context, projectID, ownerID int64, page, limit int, sortBy, sortOrder string) ([]domain.Task, int64, error)
	Update(ctx context.Context, task *domain.Task) error
	SoftDelete(ctx context.Context, id, ownerID int64) error
}
