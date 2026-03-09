package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	mw "github.com/yourusername/go-clean-auth-api/internal/http/middleware"
	"github.com/yourusername/go-clean-auth-api/internal/http/requests"
	"github.com/yourusername/go-clean-auth-api/internal/http/responses"
	"github.com/yourusername/go-clean-auth-api/internal/usecase"
)

type ProjectController struct {
	projectUsecase *usecase.ProjectUsecase
}

func NewProjectController(projectUsecase *usecase.ProjectUsecase) *ProjectController {
	return &ProjectController{projectUsecase: projectUsecase}
}

func (ctl *ProjectController) Create(c echo.Context) error {
	var req requests.CreateProjectRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	project, err := ctl.projectUsecase.Create(c.Request().Context(), usecase.CreateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusCreated, project)
}

func (ctl *ProjectController) Get(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	id, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_PROJECT_ID", err.Error(), nil)
	}

	project, err := ctl.projectUsecase.Get(c.Request().Context(), id, ownerID)
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, project)
}

func (ctl *ProjectController) List(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	page := getQueryInt(c, "page", 1)
	limit := getQueryInt(c, "limit", 20)
	sortBy := c.QueryParam("sort_by")
	sortOrder := c.QueryParam("sort_order")

	items, meta, err := ctl.projectUsecase.List(c.Request().Context(), usecase.ListProjectsInput{
		OwnerID:   ownerID,
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})
	if err != nil {
		return err
	}

	return responses.Paginated(c, http.StatusOK, items, meta)
}

func (ctl *ProjectController) Update(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	id, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_PROJECT_ID", err.Error(), nil)
	}

	var req requests.UpdateProjectRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	project, err := ctl.projectUsecase.Update(c.Request().Context(), usecase.UpdateProjectInput{
		ID:          id,
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, project)
}

func (ctl *ProjectController) Delete(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	id, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_PROJECT_ID", err.Error(), nil)
	}

	if err := ctl.projectUsecase.Delete(c.Request().Context(), id, ownerID); err != nil {
		return err
	}

	return responses.Message(c, http.StatusOK, "project deleted")
}
