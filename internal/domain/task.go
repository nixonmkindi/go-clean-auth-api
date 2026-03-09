package domain

import "time"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type Task struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	AssigneeID  *int64     `json:"assignee_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type TaskListParams struct {
	OwnerID   int64
	ProjectID *int64
	Assignee  *int64
	Status    *TaskStatus
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}
