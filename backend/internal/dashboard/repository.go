package dashboard

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Summary(ctx context.Context, sprintID *uuid.UUID) (*SummaryResponse, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Summary(ctx context.Context, sprintID *uuid.UUID) (*SummaryResponse, error) {
	query := r.db.WithContext(ctx).
		Table("tasks").
		Joins("JOIN task_statuses ON task_statuses.id = tasks.status_id").
		Where("tasks.deleted_at IS NULL")

	if sprintID != nil {
		query = query.Where("tasks.sprint_id = ?", *sprintID)
	}

	var result SummaryResponse
	err := query.Select(`
		COUNT(*) AS total_tasks,
		COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'todo' THEN 1 ELSE 0 END), 0) AS todo_tasks,
		COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'in progress' THEN 1 ELSE 0 END), 0) AS in_progress_tasks,
		COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'ready to check' THEN 1 ELSE 0 END), 0) AS ready_to_check_tasks,
		COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'checked by qa' THEN 1 ELSE 0 END), 0) AS checked_by_qa_tasks,
		COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'done' THEN 1 ELSE 0 END), 0) AS done_tasks,
		COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'blocked' THEN 1 ELSE 0 END), 0) AS blocked_tasks,
		COUNT(DISTINCT tasks.developer_id) AS total_developers,
		COUNT(DISTINCT tasks.project_id) AS total_projects
	`).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	applyCompletionRate(&result)

	return &result, nil
}

func applyCompletionRate(result *SummaryResponse) {
	if result.TotalTasks == 0 {
		result.CompletionRate = 0
		return
	}

	result.CompletionRate = (float64(result.DoneTasks) / float64(result.TotalTasks)) * 100
}
