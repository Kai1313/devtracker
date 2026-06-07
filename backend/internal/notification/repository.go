package notification

import (
	"context"
	"time"

	appquery "devtracker/backend/internal/query"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CountUnread(ctx context.Context, userID uuid.UUID, includeAll bool) (int64, error)
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID, includeAll bool) (*Notification, error)
	List(ctx context.Context, query ListQuery) ([]Notification, int64, error)
	MarkAllRead(ctx context.Context, userID uuid.UUID, includeAll bool) (int64, error)
	Update(ctx context.Context, notification *Notification) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CountUnread(ctx context.Context, userID uuid.UUID, includeAll bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("is_read = ?", false)

	if !includeAll {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Count(&count).Error

	return count, err
}

func (r *repository) Create(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID, includeAll bool) (*Notification, error) {
	var notification Notification
	query := r.db.WithContext(ctx).Where("id = ?", id)
	if !includeAll {
		query = query.Where("user_id = ?", userID)
	}

	err := query.First(&notification).Error

	return &notification, err
}

func (r *repository) List(ctx context.Context, query ListQuery) ([]Notification, int64, error) {
	dbQuery := r.db.WithContext(ctx).
		Model(&Notification{})

	if !query.IncludeAll {
		dbQuery = dbQuery.Where("user_id = ?", query.UserID)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifications []Notification
	offset := (query.Page - 1) * query.Limit
	err := dbQuery.
		Order(appquery.OrderClause(appquery.Sort{By: query.SortBy, Order: query.SortOrder}, notificationSortFields)).
		Offset(offset).
		Limit(query.Limit).
		Find(&notifications).
		Error

	return notifications, total, err
}

func (r *repository) MarkAllRead(ctx context.Context, userID uuid.UUID, includeAll bool) (int64, error) {
	now := time.Now().UTC()
	query := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("is_read = ?", false)

	if !includeAll {
		query = query.Where("user_id = ?", userID)
	}

	result := query.Updates(map[string]any{
		"is_read": true,
		"read_at": now,
	})

	return result.RowsAffected, result.Error
}

func (r *repository) Update(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}
