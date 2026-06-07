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
	TypeTaskOverdue      = "task_overdue"

	ReferenceModuleTasks = "tasks"
)

type ListQuery struct {
	Page       int
	Limit      int
	UserID     uuid.UUID
	IncludeAll bool
	SortBy     string
	SortOrder  string
}

type CreateInput struct {
	UserID          uuid.UUID
	Type            string
	Title           string
	Message         string
	ReferenceModule string
	ReferenceID     *uuid.UUID
}

type ListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}

type MarkReadResponse struct {
	Notification NotificationResponse `json:"notification"`
	UnreadCount  int64                `json:"unread_count"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

type MarkAllReadResponse struct {
	ReadCount   int64 `json:"read_count"`
	UnreadCount int64 `json:"unread_count"`
}

type NotificationResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Title           string     `json:"title"`
	Message         string     `json:"message"`
	Type            string     `json:"type"`
	ReferenceModule string     `json:"reference_module,omitempty"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	IsRead          bool       `json:"is_read"`
	ReadAt          *time.Time `json:"read_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func NewResponse(model Notification) NotificationResponse {
	return NotificationResponse{
		ID:              model.ID,
		UserID:          model.UserID,
		Title:           model.Title,
		Message:         model.Message,
		Type:            model.Type,
		ReferenceModule: model.ReferenceModule,
		ReferenceID:     model.ReferenceID,
		IsRead:          model.IsRead,
		ReadAt:          model.ReadAt,
		CreatedAt:       model.CreatedAt,
	}
}

func NewResponses(models []Notification) []NotificationResponse {
	result := make([]NotificationResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}
