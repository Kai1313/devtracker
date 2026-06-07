package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	DeveloperKPI(ctx context.Context, sprintID *uuid.UUID) ([]DeveloperKPIResponse, error)
	DeveloperSnapshotKPI(ctx context.Context, sprintID uuid.UUID) ([]DeveloperKPIResponse, error)
	GenerateSprintSnapshots(ctx context.Context, sprintID uuid.UUID) error
	ProjectKPI(ctx context.Context, sprintID *uuid.UUID) ([]ProjectKPIResponse, error)
	ProjectSnapshotKPI(ctx context.Context, sprintID uuid.UUID) ([]ProjectKPIResponse, error)
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

func (r *repository) DeveloperSnapshotKPI(ctx context.Context, sprintID uuid.UUID) ([]DeveloperKPIResponse, error) {
	var result []DeveloperKPIResponse
	err := r.db.WithContext(ctx).
		Table("kpi_snapshots").
		Joins("JOIN users ON users.id = kpi_snapshots.developer_id").
		Select(`
			kpi_snapshots.developer_id AS developer_id,
			users.name AS developer_name,
			kpi_snapshots.total_assigned AS total_assigned,
			kpi_snapshots.total_done AS total_done,
			kpi_snapshots.total_ready_to_check AS total_ready_to_check,
			kpi_snapshots.total_qa_checked AS total_checked_by_qa,
			kpi_snapshots.delayed_task_count AS delayed_tasks,
			kpi_snapshots.completion_rate AS completion_rate,
			kpi_snapshots.total_estimated_point AS total_estimated_point,
			kpi_snapshots.total_actual_point AS total_actual_point
		`).
		Where("kpi_snapshots.sprint_id = ?", sprintID).
		Order("users.name ASC").
		Scan(&result).
		Error

	return result, err
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

func (r *repository) ProjectSnapshotKPI(ctx context.Context, sprintID uuid.UUID) ([]ProjectKPIResponse, error) {
	var result []ProjectKPIResponse
	err := r.db.WithContext(ctx).
		Table("kpi_snapshots").
		Joins("JOIN sprints ON sprints.id = kpi_snapshots.sprint_id").
		Joins("JOIN projects ON projects.id = sprints.project_id").
		Select(`
			projects.id AS project_id,
			projects.project_name AS project_name,
			COALESCE(SUM(kpi_snapshots.total_assigned), 0) AS total_assigned,
			COALESCE(SUM(kpi_snapshots.total_done), 0) AS total_done,
			COALESCE(SUM(kpi_snapshots.total_ready_to_check), 0) AS total_ready_to_check,
			COALESCE(SUM(kpi_snapshots.total_qa_checked), 0) AS total_checked_by_qa,
			COALESCE(SUM(kpi_snapshots.delayed_task_count), 0) AS delayed_tasks,
			COALESCE(SUM(kpi_snapshots.total_estimated_point), 0) AS total_estimated_point,
			COALESCE(SUM(kpi_snapshots.total_actual_point), 0) AS total_actual_point
		`).
		Where("kpi_snapshots.sprint_id = ?", sprintID).
		Group("projects.id, projects.project_name").
		Order("projects.project_name ASC").
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}

	for index := range result {
		result[index].CompletionRate = completionRate(result[index].TotalDone, result[index].TotalAssigned)
	}

	return result, nil
}

func (r *repository) GenerateSprintSnapshots(ctx context.Context, sprintID uuid.UUID) error {
	var rows []DeveloperKPIResponse
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
		Where("tasks.sprint_id = ?", sprintID).
		Group("tasks.developer_id, users.name")

	if err := query.Scan(&rows).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	snapshots := make([]KPISnapshot, 0, len(rows))
	for _, row := range rows {
		snapshots = append(snapshots, KPISnapshot{
			ID:                  uuid.New(),
			DeveloperID:         row.DeveloperID,
			SprintID:            sprintID,
			TotalAssigned:       row.TotalAssigned,
			TotalDone:           row.TotalDone,
			TotalReadyToCheck:   row.TotalReadyToCheck,
			TotalQAChecked:      row.TotalCheckedByQA,
			DelayedTaskCount:    row.DelayedTasks,
			CompletionRate:      completionRate(row.TotalDone, row.TotalAssigned),
			TotalEstimatedPoint: row.TotalEstimatedPoint,
			TotalActualPoint:    row.TotalActualPoint,
			CalculatedAt:        now,
		})
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&KPISnapshot{}, "sprint_id = ?", sprintID).Error; err != nil {
			return err
		}

		if len(snapshots) == 0 {
			return nil
		}

		return tx.Create(&snapshots).Error
	})
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
