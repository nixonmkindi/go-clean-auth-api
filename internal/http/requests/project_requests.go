package requests

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"max=1000"`
}
