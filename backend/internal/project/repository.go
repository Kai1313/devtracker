package project

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByCodeIncludingDeleted(ctx context.Context, code string) (*Project, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Project, error)
	List(ctx context.Context, filter ListProjectsQuery) ([]Project, int64, error)
	Update(ctx context.Context, project *Project) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, project *Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *repository) Update(ctx context.Context, project *Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Project{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	var project Project
	err := r.db.WithContext(ctx).
		First(&project, "id = ?", id).
		Error

	return &project, err
}

func (r *repository) FindByCodeIncludingDeleted(ctx context.Context, code string) (*Project, error) {
	var project Project
	err := r.db.WithContext(ctx).
		Unscoped().
		First(&project, "LOWER(project_code) = ?", strings.ToLower(code)).
		Error

	return &project, err
}

func (r *repository) List(ctx context.Context, filter ListProjectsQuery) ([]Project, int64, error) {
	query := r.db.WithContext(ctx).Model(&Project{})

	if filter.Search != "" {
		term := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(project_code) LIKE ? OR LOWER(project_name) LIKE ? OR LOWER(client_name) LIKE ?",
			term,
			term,
			term,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var projects []Project
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&projects).
		Error

	return projects, total, err
}
