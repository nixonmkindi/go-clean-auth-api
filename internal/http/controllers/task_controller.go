package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/go-clean-auth-api/internal/domain"
	mw "github.com/yourusername/go-clean-auth-api/internal/http/middleware"
	"github.com/yourusername/go-clean-auth-api/internal/http/requests"
	"github.com/yourusername/go-clean-auth-api/internal/http/responses"
	"github.com/yourusername/go-clean-auth-api/internal/usecase"
)

type TaskController struct {
	taskUsecase *usecase.TaskUsecase
}

func NewTaskController(taskUsecase *usecase.TaskUsecase) *TaskController {
	return &TaskController{taskUsecase: taskUsecase}
}

func (ctl *TaskController) Create(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	var req requests.CreateTaskRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	dueDate, err := parseOptionalRFC3339(req.DueDate)
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_DUE_DATE", "due_date must be RFC3339", nil)
	}

	task, err := ctl.taskUsecase.Create(c.Request().Context(), usecase.CreateTaskInput{
		ProjectID:   req.ProjectID,
		OwnerID:     ownerID,
		AssigneeID:  req.AssigneeID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TaskStatus(req.Status),
		DueDate:     dueDate,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusCreated, task)
}

func (ctl *TaskController) Get(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	taskID, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_TASK_ID", err.Error(), nil)
	}

	task, err := ctl.taskUsecase.Get(c.Request().Context(), taskID, ownerID)
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, task)
}

func (ctl *TaskController) List(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	page := getQueryInt(c, "page", 1)
	limit := getQueryInt(c, "limit", 20)
	sortBy := c.QueryParam("sort_by")
	sortOrder := c.QueryParam("sort_order")

	var projectID *int64
	if raw := c.QueryParam("project_id"); raw != "" {
		id, err := mw.ParseInt64(raw, "project_id")
		if err != nil {
			return responses.Fail(c, http.StatusBadRequest, "INVALID_PROJECT_ID", err.Error(), nil)
		}
		projectID = &id
	}

	var assigneeID *int64
	if raw := c.QueryParam("assignee_id"); raw != "" {
		id, err := mw.ParseInt64(raw, "assignee_id")
		if err != nil {
			return responses.Fail(c, http.StatusBadRequest, "INVALID_ASSIGNEE_ID", err.Error(), nil)
		}
		assigneeID = &id
	}

	var status *domain.TaskStatus
	if raw := c.QueryParam("status"); raw != "" {
		s := domain.TaskStatus(raw)
		status = &s
	}

	items, meta, err := ctl.taskUsecase.List(c.Request().Context(), usecase.ListTasksInput{
		OwnerID:   ownerID,
		ProjectID: projectID,
		Assignee:  assigneeID,
		Status:    status,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		return err
	}

	return responses.Paginated(c, http.StatusOK, items, meta)
}

func (ctl *TaskController) Update(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	taskID, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_TASK_ID", err.Error(), nil)
	}

	var req requests.UpdateTaskRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	dueDate, err := parseOptionalRFC3339(req.DueDate)
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_DUE_DATE", "due_date must be RFC3339", nil)
	}

	task, err := ctl.taskUsecase.Update(c.Request().Context(), usecase.UpdateTaskInput{
		TaskID:      taskID,
		OwnerID:     ownerID,
		AssigneeID:  req.AssigneeID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TaskStatus(req.Status),
		DueDate:     dueDate,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, task)
}

func (ctl *TaskController) Assign(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	taskID, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_TASK_ID", err.Error(), nil)
	}

	var req requests.AssignTaskRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	task, err := ctl.taskUsecase.Assign(c.Request().Context(), taskID, ownerID, req.AssigneeID)
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, task)
}

func (ctl *TaskController) UpdateStatus(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	taskID, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_TASK_ID", err.Error(), nil)
	}

	var req requests.UpdateTaskStatusRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	task, err := ctl.taskUsecase.MarkStatus(c.Request().Context(), taskID, ownerID, domain.TaskStatus(req.Status))
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, task)
}

func (ctl *TaskController) Delete(c echo.Context) error {
	ownerID, err := mw.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	taskID, err := mw.ParseIDParam(c, "id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_TASK_ID", err.Error(), nil)
	}

	if err := ctl.taskUsecase.Delete(c.Request().Context(), taskID, ownerID); err != nil {
		return err
	}

	return responses.Message(c, http.StatusOK, "task deleted")
}

func (ctl *TaskController) ListByProjectForAPIClient(c echo.Context) error {
	ownerID, err := mw.GetAPIOwnerID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	projectID, err := mw.ParseIDParam(c, "project_id")
	if err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_PROJECT_ID", err.Error(), nil)
	}

	page := getQueryInt(c, "page", 1)
	limit := getQueryInt(c, "limit", 20)
	sortBy := c.QueryParam("sort_by")
	sortOrder := c.QueryParam("sort_order")

	items, meta, err := ctl.taskUsecase.ListForAPIClient(c.Request().Context(), projectID, ownerID, page, limit, sortBy, sortOrder)
	if err != nil {
		return err
	}

	return responses.Paginated(c, http.StatusOK, items, meta)
}
