package workload

import (
	"time"

	"github.com/google/uuid"
)

const (
	ClassificationLow        = "LOW"
	ClassificationNormal     = "NORMAL"
	ClassificationHigh       = "HIGH"
	ClassificationOverloaded = "OVERLOADED"
)

type Query struct {
	SprintID    string
	ProjectID   string
	DeveloperID string
	StatusID    string
	StartDate   string
	EndDate     string
}

type filter struct {
	SprintID    *uuid.UUID
	ProjectID   *uuid.UUID
	DeveloperID *uuid.UUID
	StatusID    *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	QAOnly      bool
}

type AccessScope struct {
	UserID       uuid.UUID
	IsAdmin      bool
	IsManager    bool
	IsManagement bool
	IsDeveloper  bool
	IsQA         bool
}

type DeveloperWorkloadResponse struct {
	DeveloperID          uuid.UUID `json:"developer_id"`
	DeveloperName        string    `json:"developer_name"`
	ActiveTasks          int64     `json:"active_tasks"`
	DoneTasks            int64     `json:"done_tasks"`
	OverdueTasks         int64     `json:"overdue_tasks"`
	TotalEstimatedPoints float64   `json:"total_estimated_points"`
	TotalActualPoints    float64   `json:"total_actual_points"`
	CurrentSprintTasks   int64     `json:"current_sprint_tasks"`
	WorkloadScore        int64     `json:"workload_score"`
	WorkloadLevel        string    `json:"workload_level"`
}
