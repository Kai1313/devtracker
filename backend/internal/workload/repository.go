package workload

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	DeveloperWorkload(ctx context.Context, filter filter) ([]DeveloperWorkloadResponse, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DeveloperWorkload(ctx context.Context, filter filter) ([]DeveloperWorkloadResponse, error) {
	activeCondition := "(task_statuses.is_done = FALSE AND LOWER(task_statuses.status_name) <> 'done')"
	doneCondition := "(task_statuses.is_done = TRUE OR LOWER(task_statuses.status_name) = 'done')"
	query := r.db.WithContext(ctx).
		Table("tasks").
		Joins("JOIN users ON users.id = tasks.developer_id").
		Joins("JOIN roles ON roles.id = users.role_id").
		Joins("JOIN task_statuses ON task_statuses.id = tasks.status_id").
		Joins("JOIN sprints ON sprints.id = tasks.sprint_id").
		Where("tasks.deleted_at IS NULL").
		Where("LOWER(roles.name) = ?", "developer")

	if filter.SprintID != nil {
		query = query.Where("tasks.sprint_id = ?", *filter.SprintID)
	}

	if filter.ProjectID != nil {
		query = query.Where("tasks.project_id = ?", *filter.ProjectID)
	}

	if filter.DeveloperID != nil {
		query = query.Where("tasks.developer_id = ?", *filter.DeveloperID)
	}

	if filter.StatusID != nil {
		query = query.Where("tasks.status_id = ?", *filter.StatusID)
	}

	if filter.StartDate != nil {
		query = query.Where("tasks.start_date >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("tasks.due_date <= ?", *filter.EndDate)
	}

	if filter.QAOnly {
		query = query.Where("(task_statuses.is_qa_status = TRUE OR LOWER(task_statuses.status_name) IN ?)", []string{
			"ready to check",
			"checked by qa",
		})
	}

	var result []DeveloperWorkloadResponse
	err := query.
		Select(`
			users.id AS developer_id,
			users.name AS developer_name,
			COALESCE(SUM(CASE WHEN ` + activeCondition + ` THEN 1 ELSE 0 END), 0) AS active_tasks,
			COALESCE(SUM(CASE WHEN ` + doneCondition + ` THEN 1 ELSE 0 END), 0) AS done_tasks,
			COALESCE(SUM(CASE WHEN ` + activeCondition + ` AND tasks.due_date < CURRENT_DATE THEN 1 ELSE 0 END), 0) AS overdue_tasks,
			COALESCE(SUM(tasks.estimated_point), 0) AS total_estimated_points,
			COALESCE(SUM(tasks.actual_point), 0) AS total_actual_points,
			COALESCE(SUM(CASE WHEN LOWER(sprints.status) = 'active' THEN 1 ELSE 0 END), 0) AS current_sprint_tasks
		`).
		Group("users.id, users.name").
		Order("users.name ASC").
		Scan(&result).
		Error

	if err != nil {
		return nil, err
	}

	for index := range result {
		result[index].WorkloadScore = result[index].ActiveTasks
		result[index].WorkloadLevel = classify(result[index].ActiveTasks)
	}

	return result, nil
}
