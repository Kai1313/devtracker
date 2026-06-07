package notification

import (
	"context"
	"errors"
	"strings"
	"time"

	appquery "devtracker/backend/internal/query"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repository Repository
}

var notificationSortFields = map[string]string{
	"title":      "title",
	"type":       "type",
	"is_read":    "is_read",
	"read_at":    "read_at",
	"created_at": "created_at",
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input CreateInput) error {
	if input.UserID == uuid.Nil {
		return apperrors.BadRequest("user_id is required")
	}

	notificationType := normalize(input.Type)
	if notificationType == "" {
		return apperrors.BadRequest("type is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return apperrors.BadRequest("title is required")
	}

	message := strings.TrimSpace(input.Message)
	if message == "" {
		return apperrors.BadRequest("message is required")
	}

	return s.repository.Create(ctx, &Notification{
		ID:              uuid.New(),
		UserID:          input.UserID,
		Title:           title,
		Message:         message,
		Type:            notificationType,
		ReferenceModule: normalize(input.ReferenceModule),
		ReferenceID:     input.ReferenceID,
		IsRead:          false,
	})
}

func (s *Service) List(ctx context.Context, query ListQuery) (*ListResponse, map[string]any, error) {
	query.Page = appquery.NormalizePage(query.Page)
	query.Limit = appquery.NormalizeLimit(query.Limit)

	sort, err := appquery.NormalizeSort(query.SortBy, query.SortOrder, notificationSortFields, appquery.Sort{By: "created_at", Order: appquery.Descending})
	if err != nil {
		return nil, nil, err
	}
	query.SortBy = sort.By
	query.SortOrder = sort.Order

	if query.UserID == uuid.Nil {
		return nil, nil, apperrors.BadRequest("user_id is required")
	}

	notifications, total, err := s.repository.List(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	unreadCount, err := s.repository.CountUnread(ctx, query.UserID, query.IncludeAll)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]any{
		"page":       query.Page,
		"limit":      query.Limit,
		"total":      total,
		"sort_by":    query.SortBy,
		"sort_order": query.SortOrder,
	}

	return &ListResponse{
		Notifications: NewResponses(notifications),
		UnreadCount:   unreadCount,
	}, meta, nil
}

func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID, includeAll bool) (*UnreadCountResponse, error) {
	if userID == uuid.Nil {
		return nil, apperrors.BadRequest("user_id is required")
	}

	unreadCount, err := s.repository.CountUnread(ctx, userID, includeAll)
	if err != nil {
		return nil, err
	}

	return &UnreadCountResponse{UnreadCount: unreadCount}, nil
}

func (s *Service) MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID, includeAll bool) (*MarkReadResponse, bool, error) {
	if userID == uuid.Nil {
		return nil, false, apperrors.BadRequest("user_id is required")
	}

	current, err := s.repository.FindByID(ctx, id, userID, includeAll)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, apperrors.NotFound("notification not found")
		}

		return nil, false, err
	}

	changed := false
	if !current.IsRead {
		now := time.Now().UTC()
		current.IsRead = true
		current.ReadAt = &now
		changed = true

		if err := s.repository.Update(ctx, current); err != nil {
			return nil, false, err
		}
	}

	unreadCount, err := s.repository.CountUnread(ctx, userID, includeAll)
	if err != nil {
		return nil, false, err
	}

	return &MarkReadResponse{
		Notification: NewResponse(*current),
		UnreadCount:  unreadCount,
	}, changed, nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID, includeAll bool) (*MarkAllReadResponse, error) {
	if userID == uuid.Nil {
		return nil, apperrors.BadRequest("user_id is required")
	}

	readCount, err := s.repository.MarkAllRead(ctx, userID, includeAll)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.repository.CountUnread(ctx, userID, includeAll)
	if err != nil {
		return nil, err
	}

	return &MarkAllReadResponse{
		ReadCount:   readCount,
		UnreadCount: unreadCount,
	}, nil
}

func normalize(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	return normalized
}
