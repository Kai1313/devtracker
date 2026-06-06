package kpi

import "github.com/google/uuid"

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
