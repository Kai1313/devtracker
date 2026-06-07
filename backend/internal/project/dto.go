package project

import (
	"time"

	"github.com/google/uuid"
)

type CreateProjectRequest struct {
	ProjectCode string `json:"project_code" validate:"required,min=2,max=50"`
	ProjectName string `json:"project_name" validate:"required,min=2,max=150"`
	ClientName  string `json:"client_name" validate:"omitempty,max=150"`
	Status      string `json:"status" validate:"omitempty,max=50"`
	StartDate   string `json:"start_date" validate:"omitempty"`
	EndDate     string `json:"end_date" validate:"omitempty"`
}

type UpdateProjectRequest struct {
	ProjectCode *string `json:"project_code" validate:"omitempty,min=2,max=50"`
	ProjectName *string `json:"project_name" validate:"omitempty,min=2,max=150"`
	ClientName  *string `json:"client_name" validate:"omitempty,max=150"`
	Status      *string `json:"status" validate:"omitempty,max=50"`
	StartDate   *string `json:"start_date" validate:"omitempty"`
	EndDate     *string `json:"end_date" validate:"omitempty"`
}

type ListProjectsQuery struct {
	Page      int
	Limit     int
	Search    string
	SortBy    string
	SortOrder string
}

type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	ProjectCode string    `json:"project_code"`
	ProjectName string    `json:"project_name"`
	ClientName  string    `json:"client_name,omitempty"`
	Status      string    `json:"status"`
	StartDate   *string   `json:"start_date"`
	EndDate     *string   `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewResponse(model Project) ProjectResponse {
	return ProjectResponse{
		ID:          model.ID,
		ProjectCode: model.ProjectCode,
		ProjectName: model.ProjectName,
		ClientName:  model.ClientName,
		Status:      model.Status,
		StartDate:   formatDate(model.StartDate),
		EndDate:     formatDate(model.EndDate),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func NewResponses(models []Project) []ProjectResponse {
	result := make([]ProjectResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format(dateLayout)
	return &formatted
}
