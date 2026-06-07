package workload

import "github.com/google/uuid"

const (
	ClassificationLow        = "LOW"
	ClassificationNormal     = "NORMAL"
	ClassificationHigh       = "HIGH"
	ClassificationOverloaded = "OVERLOADED"
)

type Query struct {
	SprintID  string
	ProjectID string
}

type filter struct {
	SprintID  *uuid.UUID
	ProjectID *uuid.UUID
}

type DeveloperWorkloadResponse struct {
	DeveloperID            uuid.UUID `json:"developer_id"`
	DeveloperName          string    `json:"developer_name"`
	ActiveTasks            int64     `json:"active_tasks"`
	TotalPoints            float64   `json:"total_points"`
	OverdueTasks           int64     `json:"overdue_tasks"`
	CurrentSprintTasks     int64     `json:"current_sprint_tasks"`
	WorkloadClassification string    `json:"workload_classification"`
}
