package task

import (
	"time"

	"devtracker/backend/internal/project"
	"devtracker/backend/internal/sprint"
	"devtracker/backend/internal/status"
	"devtracker/backend/internal/user"

	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	DeveloperID     string   `json:"developer_id" validate:"required,uuid"`
	ProjectID       string   `json:"project_id" validate:"required,uuid"`
	SprintID        string   `json:"sprint_id" validate:"required,uuid"`
	TicketNumber    string   `json:"ticket_number" validate:"omitempty,max=100"`
	TaskTitle       string   `json:"task_title" validate:"required,min=2,max=255"`
	TaskDescription string   `json:"task_description" validate:"omitempty"`
	Priority        string   `json:"priority" validate:"required,max=50"`
	StatusID        string   `json:"status_id" validate:"required,uuid"`
	EstimatedPoint  *float64 `json:"estimated_point" validate:"omitempty,gte=0"`
	ActualPoint     *float64 `json:"actual_point" validate:"omitempty,gte=0"`
	StartDate       string   `json:"start_date" validate:"omitempty"`
	DueDate         string   `json:"due_date" validate:"omitempty"`
	CompletedDate   string   `json:"completed_date" validate:"omitempty"`
	QACheckedDate   string   `json:"qa_checked_date" validate:"omitempty"`
}

type UpdateTaskRequest struct {
	DeveloperID     *string  `json:"developer_id" validate:"omitempty,uuid"`
	ProjectID       *string  `json:"project_id" validate:"omitempty,uuid"`
	SprintID        *string  `json:"sprint_id" validate:"omitempty,uuid"`
	TicketNumber    *string  `json:"ticket_number" validate:"omitempty,max=100"`
	TaskTitle       *string  `json:"task_title" validate:"omitempty,min=2,max=255"`
	TaskDescription *string  `json:"task_description" validate:"omitempty"`
	Priority        *string  `json:"priority" validate:"omitempty,max=50"`
	StatusID        *string  `json:"status_id" validate:"omitempty,uuid"`
	EstimatedPoint  *float64 `json:"estimated_point" validate:"omitempty,gte=0"`
	ActualPoint     *float64 `json:"actual_point" validate:"omitempty,gte=0"`
	StartDate       *string  `json:"start_date" validate:"omitempty"`
	DueDate         *string  `json:"due_date" validate:"omitempty"`
	CompletedDate   *string  `json:"completed_date" validate:"omitempty"`
	QACheckedDate   *string  `json:"qa_checked_date" validate:"omitempty"`
}

type ListTasksQuery struct {
	Page        int
	Limit       int
	DeveloperID string
	ProjectID   string
	SprintID    string
	StatusID    string
	Search      string
}

type TaskResponse struct {
	ID              uuid.UUID                 `json:"id"`
	DeveloperID     uuid.UUID                 `json:"developer_id"`
	Developer       user.UserResponse         `json:"developer"`
	ProjectID       uuid.UUID                 `json:"project_id"`
	Project         project.ProjectResponse   `json:"project"`
	SprintID        uuid.UUID                 `json:"sprint_id"`
	Sprint          sprint.SprintResponse     `json:"sprint"`
	StatusID        uuid.UUID                 `json:"status_id"`
	Status          status.TaskStatusResponse `json:"status"`
	TicketNumber    string                    `json:"ticket_number,omitempty"`
	TaskTitle       string                    `json:"task_title"`
	TaskDescription string                    `json:"task_description,omitempty"`
	Priority        string                    `json:"priority"`
	EstimatedPoint  *float64                  `json:"estimated_point"`
	ActualPoint     *float64                  `json:"actual_point"`
	StartDate       *string                   `json:"start_date"`
	DueDate         *string                   `json:"due_date"`
	CompletedDate   *string                   `json:"completed_date"`
	QACheckedDate   *string                   `json:"qa_checked_date"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

func NewResponse(model Task) TaskResponse {
	return TaskResponse{
		ID:              model.ID,
		DeveloperID:     model.DeveloperID,
		Developer:       user.NewResponse(model.Developer),
		ProjectID:       model.ProjectID,
		Project:         project.NewResponse(model.Project),
		SprintID:        model.SprintID,
		Sprint:          sprint.NewResponse(model.Sprint),
		StatusID:        model.StatusID,
		Status:          status.NewResponse(model.Status),
		TicketNumber:    model.TicketNumber,
		TaskTitle:       model.TaskTitle,
		TaskDescription: model.TaskDescription,
		Priority:        model.Priority,
		EstimatedPoint:  model.EstimatedPoint,
		ActualPoint:     model.ActualPoint,
		StartDate:       formatDate(model.StartDate),
		DueDate:         formatDate(model.DueDate),
		CompletedDate:   formatTimestamp(model.CompletedDate),
		QACheckedDate:   formatTimestamp(model.QACheckedDate),
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

func NewResponses(models []Task) []TaskResponse {
	result := make([]TaskResponse, 0, len(models))
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

func formatTimestamp(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
