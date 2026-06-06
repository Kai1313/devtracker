package sprint

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, sprint *Sprint) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Sprint, error)
	List(ctx context.Context, filter ListSprintsQuery) ([]Sprint, int64, error)
	Update(ctx context.Context, sprint *Sprint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, sprint *Sprint) error {
	return r.db.WithContext(ctx).Omit("Project").Create(sprint).Error
}

func (r *repository) Update(ctx context.Context, sprint *Sprint) error {
	return r.db.WithContext(ctx).Omit("Project").Save(sprint).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Sprint{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Sprint, error) {
	var sprint Sprint
	err := r.db.WithContext(ctx).
		Preload("Project").
		First(&sprint, "id = ?", id).
		Error

	return &sprint, err
}

func (r *repository) List(ctx context.Context, filter ListSprintsQuery) ([]Sprint, int64, error) {
	query := r.db.WithContext(ctx).Model(&Sprint{}).Preload("Project")

	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var sprints []Sprint
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Order("start_date DESC").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&sprints).
		Error

	return sprints, total, err
}
