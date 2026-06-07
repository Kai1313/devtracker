package audit

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter listFilter) ([]AuditLog, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) List(ctx context.Context, filter listFilter) ([]AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&AuditLog{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.Module != "" {
		query = query.Where("module = ?", filter.Module)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at < ?", *filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []AuditLog
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&logs).
		Error

	return logs, total, err
}
