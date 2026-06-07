package notification

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	Create(ctx context.Context, notification *Notification) error
	FindByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Notification, error)
	List(ctx context.Context, query ListQuery) ([]Notification, int64, error)
	Update(ctx context.Context, notification *Notification) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).
		Error

	return count, err
}

func (r *repository) Create(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *repository) FindByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Notification, error) {
	var notification Notification
	err := r.db.WithContext(ctx).
		First(&notification, "id = ? AND user_id = ?", id, userID).
		Error

	return &notification, err
}

func (r *repository) List(ctx context.Context, query ListQuery) ([]Notification, int64, error) {
	dbQuery := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ?", query.UserID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifications []Notification
	offset := (query.Page - 1) * query.Limit
	err := dbQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(query.Limit).
		Find(&notifications).
		Error

	return notifications, total, err
}

func (r *repository) Update(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}
