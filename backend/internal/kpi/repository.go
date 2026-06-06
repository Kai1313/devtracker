package kpi

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	DeveloperKPI(ctx context.Context, sprintID *uuid.UUID) ([]DeveloperKPIResponse, error)
	ProjectKPI(ctx context.Context, sprintID *uuid.UUID) ([]ProjectKPIResponse, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DeveloperKPI(ctx context.Context, sprintID *uuid.UUID) ([]DeveloperKPIResponse, error) {
	query := baseTaskKPIQuery(r.db.WithContext(ctx)).
		Joins("JOIN users ON users.id = tasks.developer_id").
		Select(`
			tasks.developer_id AS developer_id,
			users.name AS developer_name,
			COUNT(*) AS total_assigned,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'done' THEN 1 ELSE 0 END), 0) AS total_done,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'ready to check' THEN 1 ELSE 0 END), 0) AS total_ready_to_check,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'checked by qa' THEN 1 ELSE 0 END), 0) AS total_checked_by_qa,
			COALESCE(SUM(CASE WHEN tasks.due_date < CURRENT_DATE AND LOWER(task_statuses.status_name) <> 'done' THEN 1 ELSE 0 END), 0) AS delayed_tasks,
			COALESCE(SUM(tasks.estimated_point), 0) AS total_estimated_point,
			COALESCE(SUM(tasks.actual_point), 0) AS total_actual_point
		`).
		Group("tasks.developer_id, users.name").
		Order("users.name ASC")

	if sprintID != nil {
		query = query.Where("tasks.sprint_id = ?", *sprintID)
	}

	var result []DeveloperKPIResponse
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}

	for index := range result {
		result[index].CompletionRate = completionRate(result[index].TotalDone, result[index].TotalAssigned)
	}

	return result, nil
}

func (r *repository) ProjectKPI(ctx context.Context, sprintID *uuid.UUID) ([]ProjectKPIResponse, error) {
	query := baseTaskKPIQuery(r.db.WithContext(ctx)).
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Select(`
			tasks.project_id AS project_id,
			projects.project_name AS project_name,
			COUNT(*) AS total_assigned,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'done' THEN 1 ELSE 0 END), 0) AS total_done,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'ready to check' THEN 1 ELSE 0 END), 0) AS total_ready_to_check,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'checked by qa' THEN 1 ELSE 0 END), 0) AS total_checked_by_qa,
			COALESCE(SUM(CASE WHEN tasks.due_date < CURRENT_DATE AND LOWER(task_statuses.status_name) <> 'done' THEN 1 ELSE 0 END), 0) AS delayed_tasks,
			COALESCE(SUM(tasks.estimated_point), 0) AS total_estimated_point,
			COALESCE(SUM(tasks.actual_point), 0) AS total_actual_point
		`).
		Group("tasks.project_id, projects.project_name").
		Order("projects.project_name ASC")

	if sprintID != nil {
		query = query.Where("tasks.sprint_id = ?", *sprintID)
	}

	var result []ProjectKPIResponse
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}

	for index := range result {
		result[index].CompletionRate = completionRate(result[index].TotalDone, result[index].TotalAssigned)
	}

	return result, nil
}

func baseTaskKPIQuery(db *gorm.DB) *gorm.DB {
	return db.
		Table("tasks").
		Joins("JOIN task_statuses ON task_statuses.id = tasks.status_id").
		Where("tasks.deleted_at IS NULL")
}

func completionRate(done int64, total int64) float64 {
	if total == 0 {
		return 0
	}

	return (float64(done) / float64(total)) * 100
}
