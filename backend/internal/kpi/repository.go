package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	DeveloperKPI(ctx context.Context, sprintID *uuid.UUID) ([]DeveloperKPIResponse, error)
	DeveloperSnapshotKPI(ctx context.Context, sprintID uuid.UUID) ([]DeveloperKPIResponse, error)
	DeveloperSnapshots(ctx context.Context, developerID uuid.UUID) ([]KPISnapshotResponse, error)
	GenerateSprintSnapshots(ctx context.Context, sprintID uuid.UUID) ([]KPISnapshotResponse, error)
	ListSnapshots(ctx context.Context, sprintID *uuid.UUID, developerID *uuid.UUID) ([]KPISnapshotResponse, error)
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
			kpi_snapshots.total_assigned_tasks AS total_assigned,
			kpi_snapshots.total_done_tasks AS total_done,
			kpi_snapshots.total_ready_to_check_tasks AS total_ready_to_check,
			kpi_snapshots.total_checked_by_qa_tasks AS total_checked_by_qa,
			kpi_snapshots.delayed_tasks AS delayed_tasks,
			kpi_snapshots.completion_rate AS completion_rate,
			kpi_snapshots.total_estimated_points AS total_estimated_point,
			kpi_snapshots.total_actual_points AS total_actual_point
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
			COALESCE(SUM(kpi_snapshots.total_assigned_tasks), 0) AS total_assigned,
			COALESCE(SUM(kpi_snapshots.total_done_tasks), 0) AS total_done,
			COALESCE(SUM(kpi_snapshots.total_ready_to_check_tasks), 0) AS total_ready_to_check,
			COALESCE(SUM(kpi_snapshots.total_checked_by_qa_tasks), 0) AS total_checked_by_qa,
			COALESCE(SUM(kpi_snapshots.delayed_tasks), 0) AS delayed_tasks,
			COALESCE(SUM(kpi_snapshots.total_estimated_points), 0) AS total_estimated_point,
			COALESCE(SUM(kpi_snapshots.total_actual_points), 0) AS total_actual_point
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

func (r *repository) ListSnapshots(ctx context.Context, sprintID *uuid.UUID, developerID *uuid.UUID) ([]KPISnapshotResponse, error) {
	query := r.snapshotResponseQuery(ctx)

	if sprintID != nil {
		query = query.Where("kpi_snapshots.sprint_id = ?", *sprintID)
	}

	if developerID != nil {
		query = query.Where("kpi_snapshots.developer_id = ?", *developerID)
	}

	var result []KPISnapshotResponse
	err := query.
		Order("kpi_snapshots.generated_at DESC").
		Order("users.name ASC").
		Scan(&result).
		Error

	return result, err
}

func (r *repository) DeveloperSnapshots(ctx context.Context, developerID uuid.UUID) ([]KPISnapshotResponse, error) {
	var result []KPISnapshotResponse
	err := r.snapshotResponseQuery(ctx).
		Where("kpi_snapshots.developer_id = ?", developerID).
		Order("kpi_snapshots.generated_at DESC").
		Scan(&result).
		Error

	return result, err
}

func (r *repository) GenerateSprintSnapshots(ctx context.Context, sprintID uuid.UUID) ([]KPISnapshotResponse, error) {
	var rows []snapshotMetricRow
	query := baseTaskKPIQuery(r.db.WithContext(ctx)).
		Joins("JOIN users ON users.id = tasks.developer_id").
		Select(`
			tasks.developer_id AS developer_id,
			users.name AS developer_name,
			COUNT(*) AS total_assigned_tasks,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'done' THEN 1 ELSE 0 END), 0) AS total_done_tasks,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'ready to check' THEN 1 ELSE 0 END), 0) AS total_ready_to_check_tasks,
			COALESCE(SUM(CASE WHEN LOWER(task_statuses.status_name) = 'checked by qa' THEN 1 ELSE 0 END), 0) AS total_checked_by_qa_tasks,
			COALESCE(SUM(CASE WHEN tasks.due_date < CURRENT_DATE AND LOWER(task_statuses.status_name) <> 'done' THEN 1 ELSE 0 END), 0) AS delayed_tasks,
			COALESCE(SUM(tasks.estimated_point), 0) AS total_estimated_points,
			COALESCE(SUM(tasks.actual_point), 0) AS total_actual_points,
			COALESCE(AVG(CASE
				WHEN tasks.completed_date IS NOT NULL AND tasks.start_date IS NOT NULL
				THEN EXTRACT(EPOCH FROM (tasks.completed_date - tasks.start_date::timestamp)) / 3600
				ELSE NULL
			END), 0) AS average_completion_time_hours
		`).
		Where("tasks.sprint_id = ?", sprintID).
		Group("tasks.developer_id, users.name")

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	snapshots := make([]KPISnapshot, 0, len(rows))
	for _, row := range rows {
		snapshots = append(snapshots, KPISnapshot{
			ID:                         uuid.New(),
			SprintID:                   sprintID,
			DeveloperID:                row.DeveloperID,
			TotalAssignedTasks:         row.TotalAssignedTasks,
			TotalDoneTasks:             row.TotalDoneTasks,
			TotalReadyToCheckTasks:     row.TotalReadyToCheckTasks,
			TotalCheckedByQATasks:      row.TotalCheckedByQATasks,
			DelayedTasks:               row.DelayedTasks,
			CompletionRate:             completionRate(row.TotalDoneTasks, row.TotalAssignedTasks),
			TotalEstimatedPoints:       row.TotalEstimatedPoints,
			TotalActualPoints:          row.TotalActualPoints,
			AverageCompletionTimeHours: row.AverageCompletionTimeHours,
			GeneratedAt:                now,
		})
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(snapshots) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "sprint_id"},
				{Name: "developer_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"total_assigned_tasks",
				"total_done_tasks",
				"total_ready_to_check_tasks",
				"total_checked_by_qa_tasks",
				"delayed_tasks",
				"completion_rate",
				"total_estimated_points",
				"total_actual_points",
				"average_completion_time_hours",
				"generated_at",
			}),
		}).Create(&snapshots).Error
	})
	if err != nil {
		return nil, err
	}

	return r.ListSnapshots(ctx, &sprintID, nil)
}

func (r *repository) snapshotResponseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("kpi_snapshots").
		Joins("JOIN users ON users.id = kpi_snapshots.developer_id").
		Select(`
			kpi_snapshots.id AS id,
			kpi_snapshots.sprint_id AS sprint_id,
			kpi_snapshots.developer_id AS developer_id,
			users.name AS developer_name,
			kpi_snapshots.total_assigned_tasks AS total_assigned_tasks,
			kpi_snapshots.total_done_tasks AS total_done_tasks,
			kpi_snapshots.total_ready_to_check_tasks AS total_ready_to_check_tasks,
			kpi_snapshots.total_checked_by_qa_tasks AS total_checked_by_qa_tasks,
			kpi_snapshots.delayed_tasks AS delayed_tasks,
			kpi_snapshots.completion_rate AS completion_rate,
			kpi_snapshots.total_estimated_points AS total_estimated_points,
			kpi_snapshots.total_actual_points AS total_actual_points,
			kpi_snapshots.average_completion_time_hours AS average_completion_time_hours,
			kpi_snapshots.generated_at AS generated_at,
			kpi_snapshots.created_at AS created_at
		`)
}

type snapshotMetricRow struct {
	DeveloperID                uuid.UUID
	DeveloperName              string
	TotalAssignedTasks         int64
	TotalDoneTasks             int64
	TotalReadyToCheckTasks     int64
	TotalCheckedByQATasks      int64
	DelayedTasks               int64
	TotalEstimatedPoints       float64
	TotalActualPoints          float64
	AverageCompletionTimeHours float64
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
