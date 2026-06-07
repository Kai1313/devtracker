package task

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, task *Task, history *TaskHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Task, error)
	List(ctx context.Context, filter ListTasksQuery) ([]Task, int64, error)
	ListWithAccess(ctx context.Context, filter ListTasksQuery, access ListAccessFilter) ([]Task, int64, error)
	ListHistories(ctx context.Context, taskID uuid.UUID) ([]TaskHistory, error)
	Update(ctx context.Context, task *Task, history *TaskHistory) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, task *Task, history *TaskHistory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Omit("Developer", "Project", "Sprint", "Status").
			Create(task).
			Error; err != nil {
			return err
		}

		if history == nil {
			return nil
		}

		history.TaskID = task.ID
		return tx.Create(history).Error
	})
}

func (r *repository) Update(ctx context.Context, task *Task, history *TaskHistory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Omit("Developer", "Project", "Sprint", "Status").
			Save(task).
			Error; err != nil {
			return err
		}

		if history == nil {
			return nil
		}

		history.TaskID = task.ID
		return tx.Create(history).Error
	})
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Task{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	var task Task
	err := withTaskPreloads(r.db.WithContext(ctx)).
		First(&task, "id = ?", id).
		Error

	return &task, err
}

func (r *repository) List(ctx context.Context, filter ListTasksQuery) ([]Task, int64, error) {
	return r.list(ctx, filter, ListAccessFilter{})
}

func (r *repository) ListWithAccess(ctx context.Context, filter ListTasksQuery, access ListAccessFilter) ([]Task, int64, error) {
	return r.list(ctx, filter, access)
}

func (r *repository) list(ctx context.Context, filter ListTasksQuery, access ListAccessFilter) ([]Task, int64, error) {
	query := withTaskPreloads(r.db.WithContext(ctx).Model(&Task{}))

	if !access.IsZero() {
		query = applyAccessFilter(query, access)
	}

	if filter.DeveloperID != "" {
		query = query.Where("developer_id = ?", filter.DeveloperID)
	}

	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if filter.SprintID != "" {
		query = query.Where("sprint_id = ?", filter.SprintID)
	}

	if filter.StatusID != "" {
		query = query.Where("status_id = ?", filter.StatusID)
	}

	if filter.Search != "" {
		term := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(ticket_number) LIKE ? OR LOWER(task_title) LIKE ?", term, term)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []Task
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&tasks).
		Error

	return tasks, total, err
}

func applyAccessFilter(query *gorm.DB, access ListAccessFilter) *gorm.DB {
	if access.DeveloperID != uuid.Nil && access.ReadyToCheckStatusID != uuid.Nil {
		return query.Where(
			"(developer_id = ? OR status_id = ?)",
			access.DeveloperID,
			access.ReadyToCheckStatusID,
		)
	}

	if access.DeveloperID != uuid.Nil {
		return query.Where("developer_id = ?", access.DeveloperID)
	}

	return query.Where("status_id = ?", access.ReadyToCheckStatusID)
}

func (r *repository) ListHistories(ctx context.Context, taskID uuid.UUID) ([]TaskHistory, error) {
	var histories []TaskHistory
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("changed_at DESC").
		Find(&histories).
		Error

	return histories, err
}

func withTaskPreloads(query *gorm.DB) *gorm.DB {
	return query.
		Preload("Developer.Role").
		Preload("Project").
		Preload("Sprint.Project").
		Preload("Status")
}
