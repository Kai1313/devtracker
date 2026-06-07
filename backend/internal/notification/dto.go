package notification

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeTaskAssigned     = "task_assigned"
	TypeTaskReadyToCheck = "task_ready_to_check"
	TypeTaskCheckedByQA  = "task_checked_by_qa"
	TypeTaskDone         = "task_done"
)

type ListQuery struct {
	Page   int
	Limit  int
	UserID uuid.UUID
}

type CreateInput struct {
	UserID  uuid.UUID
	TaskID  *uuid.UUID
	Type    string
	Title   string
	Message string
}

type ListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}

type MarkReadResponse struct {
	Notification NotificationResponse `json:"notification"`
	UnreadCount  int64                `json:"unread_count"`
}

type NotificationResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TaskID    *uuid.UUID `json:"task_id,omitempty"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewResponse(model Notification) NotificationResponse {
	return NotificationResponse{
		ID:        model.ID,
		UserID:    model.UserID,
		TaskID:    model.TaskID,
		Type:      model.Type,
		Title:     model.Title,
		Message:   model.Message,
		IsRead:    model.IsRead,
		ReadAt:    model.ReadAt,
		CreatedAt: model.CreatedAt,
	}
}

func NewResponses(models []Notification) []NotificationResponse {
	result := make([]NotificationResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}
