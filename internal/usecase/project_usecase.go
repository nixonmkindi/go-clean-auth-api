package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
	"github.com/nixonmkindi/go-clean-auth-api/internal/pkg/apperror"
	"github.com/nixonmkindi/go-clean-auth-api/internal/repository"
)

type ProjectUsecase struct {
	repo         repository.ProjectRepository
	defaultLimit int
	maxLimit     int
}

type CreateProjectInput struct {
	Name        string
	Description string
	OwnerID     int64
}

type UpdateProjectInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
}

type ListProjectsInput struct {
	OwnerID   int64
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}

func NewProjectUsecase(repo repository.ProjectRepository, defaultLimit, maxLimit int) *ProjectUsecase {
	return &ProjectUsecase{repo: repo, defaultLimit: defaultLimit, maxLimit: maxLimit}
}

func (u *ProjectUsecase) Create(ctx context.Context, input CreateProjectInput) (*domain.Project, error) {
	project := &domain.Project{
		OwnerID:     input.OwnerID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
	}

	if err := u.repo.Create(ctx, project); err != nil {
		return nil, apperror.New(500, "PROJECT_CREATE_ERROR", "failed to create project", nil, err)
	}

	return project, nil
}

func (u *ProjectUsecase) Get(ctx context.Context, id, ownerID int64) (*domain.Project, error) {
	project, err := u.repo.GetByIDForOwner(ctx, id, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(404, "PROJECT_NOT_FOUND", "project not found", nil, err)
		}
		return nil, apperror.New(500, "PROJECT_FETCH_ERROR", "failed to fetch project", nil, err)
	}
	return project, nil
}

func (u *ProjectUsecase) List(ctx context.Context, input ListProjectsInput) ([]domain.Project, PaginationMeta, error) {
	page, limit := NormalizePagination(input.Page, input.Limit, u.defaultLimit, u.maxLimit)

	items, total, err := u.repo.ListByOwner(ctx, repository.ProjectListParams{
		OwnerID:   input.OwnerID,
		Page:      page,
		Limit:     limit,
		SortBy:    input.SortBy,
		SortOrder: input.SortOrder,
	})
	if err != nil {
		return nil, PaginationMeta{}, apperror.New(500, "PROJECT_LIST_ERROR", "failed to list projects", nil, err)
	}

	return items, BuildPaginationMeta(page, limit, total), nil
}

func (u *ProjectUsecase) Update(ctx context.Context, input UpdateProjectInput) (*domain.Project, error) {
	project, err := u.Get(ctx, input.ID, input.OwnerID)
	if err != nil {
		return nil, err
	}

	project.Name = strings.TrimSpace(input.Name)
	project.Description = strings.TrimSpace(input.Description)

	if err := u.repo.Update(ctx, project); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(404, "PROJECT_NOT_FOUND", "project not found", nil, err)
		}
		return nil, apperror.New(500, "PROJECT_UPDATE_ERROR", "failed to update project", nil, err)
	}

	return project, nil
}

func (u *ProjectUsecase) Delete(ctx context.Context, id, ownerID int64) error {
	if err := u.repo.SoftDelete(ctx, id, ownerID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return apperror.New(404, "PROJECT_NOT_FOUND", "project not found", nil, err)
		}
		return apperror.New(500, "PROJECT_DELETE_ERROR", "failed to delete project", nil, err)
	}
	return nil
}
