package status

import (
	"context"
	"strings"

	appquery "devtracker/backend/internal/query"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, taskStatus *TaskStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*TaskStatus, error)
	FindByName(ctx context.Context, name string) (*TaskStatus, error)
	List(ctx context.Context, filter ListTaskStatusesQuery) ([]TaskStatus, int64, error)
	Update(ctx context.Context, taskStatus *TaskStatus) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, taskStatus *TaskStatus) error {
	return r.db.WithContext(ctx).Create(taskStatus).Error
}

func (r *repository) Update(ctx context.Context, taskStatus *TaskStatus) error {
	return r.db.WithContext(ctx).Save(taskStatus).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&TaskStatus{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*TaskStatus, error) {
	var taskStatus TaskStatus
	err := r.db.WithContext(ctx).
		First(&taskStatus, "id = ?", id).
		Error

	return &taskStatus, err
}

func (r *repository) FindByName(ctx context.Context, name string) (*TaskStatus, error) {
	var taskStatus TaskStatus
	err := r.db.WithContext(ctx).
		First(&taskStatus, "LOWER(status_name) = ?", strings.ToLower(name)).
		Error

	return &taskStatus, err
}

func (r *repository) List(ctx context.Context, filter ListTaskStatusesQuery) ([]TaskStatus, int64, error) {
	query := r.db.WithContext(ctx).Model(&TaskStatus{})

	if filter.Search != "" {
		term := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(status_name) LIKE ? OR LOWER(color_name) LIKE ? OR LOWER(color_hex) LIKE ?",
			term,
			term,
			term,
		)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var statuses []TaskStatus
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Order(appquery.OrderClause(appquery.Sort{By: filter.SortBy, Order: filter.SortOrder}, taskStatusSortFields)).
		Order("created_at ASC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&statuses).
		Error

	return statuses, total, err
}
