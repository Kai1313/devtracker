package sprint

import (
	"time"

	"devtracker/backend/internal/project"

	"github.com/google/uuid"
)

type CreateSprintRequest struct {
	ProjectID  string `json:"project_id" validate:"required,uuid"`
	SprintName string `json:"sprint_name" validate:"required,min=2,max=150"`
	StartDate  string `json:"start_date" validate:"required"`
	EndDate    string `json:"end_date" validate:"required"`
	Status     string `json:"status" validate:"omitempty,max=50"`
}

type UpdateSprintRequest struct {
	ProjectID  *string `json:"project_id" validate:"omitempty,uuid"`
	SprintName *string `json:"sprint_name" validate:"omitempty,min=2,max=150"`
	StartDate  *string `json:"start_date" validate:"omitempty"`
	EndDate    *string `json:"end_date" validate:"omitempty"`
	Status     *string `json:"status" validate:"omitempty,max=50"`
}

type ListSprintsQuery struct {
	Page      int
	Limit     int
	ProjectID string
	Status    string
	SortBy    string
	SortOrder string
}

type SprintResponse struct {
	ID         uuid.UUID               `json:"id"`
	ProjectID  uuid.UUID               `json:"project_id"`
	Project    project.ProjectResponse `json:"project"`
	SprintName string                  `json:"sprint_name"`
	StartDate  string                  `json:"start_date"`
	EndDate    string                  `json:"end_date"`
	Status     string                  `json:"status"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}

func NewResponse(model Sprint) SprintResponse {
	return SprintResponse{
		ID:         model.ID,
		ProjectID:  model.ProjectID,
		Project:    project.NewResponse(model.Project),
		SprintName: model.SprintName,
		StartDate:  model.StartDate.Format(dateLayout),
		EndDate:    model.EndDate.Format(dateLayout),
		Status:     model.Status,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

func NewResponses(models []Sprint) []SprintResponse {
	result := make([]SprintResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}
