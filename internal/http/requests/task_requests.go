package requests

type CreateTaskRequest struct {
	ProjectID   int64   `json:"project_id" validate:"required,gt=0"`
	AssigneeID  *int64  `json:"assignee_id,omitempty" validate:"omitempty,gt=0"`
	Title       string  `json:"title" validate:"required,min=3,max=120"`
	Description string  `json:"description" validate:"max=1500"`
	Status      string  `json:"status" validate:"required,oneof=pending in_progress done"`
	DueDate     *string `json:"due_date,omitempty"`
}

type UpdateTaskRequest struct {
	AssigneeID  *int64  `json:"assignee_id,omitempty" validate:"omitempty,gt=0"`
	Title       string  `json:"title" validate:"required,min=3,max=120"`
	Description string  `json:"description" validate:"max=1500"`
	Status      string  `json:"status" validate:"required,oneof=pending in_progress done"`
	DueDate     *string `json:"due_date,omitempty"`
}

type AssignTaskRequest struct {
	AssigneeID int64 `json:"assignee_id" validate:"required,gt=0"`
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending in_progress done"`
}
