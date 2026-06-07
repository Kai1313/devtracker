package kpi

import (
	"time"

	"github.com/google/uuid"
)

type Query struct {
	SprintID string
}

type DeveloperKPIResponse struct {
	DeveloperID         uuid.UUID `json:"developer_id"`
	DeveloperName       string    `json:"developer_name"`
	TotalAssigned       int64     `json:"total_assigned"`
	TotalDone           int64     `json:"total_done"`
	TotalReadyToCheck   int64     `json:"total_ready_to_check"`
	TotalCheckedByQA    int64     `json:"total_checked_by_qa"`
	DelayedTasks        int64     `json:"delayed_tasks"`
	CompletionRate      float64   `json:"completion_rate"`
	TotalEstimatedPoint float64   `json:"total_estimated_point"`
	TotalActualPoint    float64   `json:"total_actual_point"`
}

type ProjectKPIResponse struct {
	ProjectID           uuid.UUID `json:"project_id"`
	ProjectName         string    `json:"project_name"`
	TotalAssigned       int64     `json:"total_assigned"`
	TotalDone           int64     `json:"total_done"`
	TotalReadyToCheck   int64     `json:"total_ready_to_check"`
	TotalCheckedByQA    int64     `json:"total_checked_by_qa"`
	DelayedTasks        int64     `json:"delayed_tasks"`
	CompletionRate      float64   `json:"completion_rate"`
	TotalEstimatedPoint float64   `json:"total_estimated_point"`
	TotalActualPoint    float64   `json:"total_actual_point"`
}

type SnapshotQuery struct {
	SprintID string
}

type KPISnapshotResponse struct {
	ID                         uuid.UUID `json:"id"`
	SprintID                   uuid.UUID `json:"sprint_id"`
	DeveloperID                uuid.UUID `json:"developer_id"`
	DeveloperName              string    `json:"developer_name,omitempty"`
	TotalAssignedTasks         int64     `json:"total_assigned_tasks"`
	TotalDoneTasks             int64     `json:"total_done_tasks"`
	TotalReadyToCheckTasks     int64     `json:"total_ready_to_check_tasks"`
	TotalCheckedByQATasks      int64     `json:"total_checked_by_qa_tasks"`
	DelayedTasks               int64     `json:"delayed_tasks"`
	CompletionRate             float64   `json:"completion_rate"`
	TotalEstimatedPoints       float64   `json:"total_estimated_points"`
	TotalActualPoints          float64   `json:"total_actual_points"`
	AverageCompletionTimeHours float64   `json:"average_completion_time_hours"`
	GeneratedAt                time.Time `json:"generated_at"`
	CreatedAt                  time.Time `json:"created_at"`
}

type SnapshotScope struct {
	UserID       uuid.UUID
	IsAdmin      bool
	IsManager    bool
	IsManagement bool
	IsDeveloper  bool
}

func NewSnapshotResponse(model KPISnapshot) KPISnapshotResponse {
	return KPISnapshotResponse{
		ID:                         model.ID,
		SprintID:                   model.SprintID,
		DeveloperID:                model.DeveloperID,
		TotalAssignedTasks:         model.TotalAssignedTasks,
		TotalDoneTasks:             model.TotalDoneTasks,
		TotalReadyToCheckTasks:     model.TotalReadyToCheckTasks,
		TotalCheckedByQATasks:      model.TotalCheckedByQATasks,
		DelayedTasks:               model.DelayedTasks,
		CompletionRate:             model.CompletionRate,
		TotalEstimatedPoints:       model.TotalEstimatedPoints,
		TotalActualPoints:          model.TotalActualPoints,
		AverageCompletionTimeHours: model.AverageCompletionTimeHours,
		GeneratedAt:                model.GeneratedAt,
		CreatedAt:                  model.CreatedAt,
	}
}
