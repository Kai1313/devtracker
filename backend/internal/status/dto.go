package status

import (
	"time"

	"github.com/google/uuid"
)

type CreateTaskStatusRequest struct {
	StatusName  string `json:"status_name" validate:"required,min=2,max=100"`
	ColorName   string `json:"color_name" validate:"required,min=2,max=30"`
	ColorHex    string `json:"color_hex" validate:"required,max=7"`
	StatusOrder int    `json:"status_order" validate:"min=0"`
	IsDone      bool   `json:"is_done"`
	IsQAStatus  bool   `json:"is_qa_status"`
	IsActive    *bool  `json:"is_active" validate:"omitempty"`
}

type UpdateTaskStatusRequest struct {
	StatusName  *string `json:"status_name" validate:"omitempty,min=2,max=100"`
	ColorName   *string `json:"color_name" validate:"omitempty,min=2,max=30"`
	ColorHex    *string `json:"color_hex" validate:"omitempty,max=7"`
	StatusOrder *int    `json:"status_order" validate:"omitempty,min=0"`
	IsDone      *bool   `json:"is_done" validate:"omitempty"`
	IsQAStatus  *bool   `json:"is_qa_status" validate:"omitempty"`
	IsActive    *bool   `json:"is_active" validate:"omitempty"`
}

type ListTaskStatusesQuery struct {
	Page     int
	Limit    int
	IsActive *bool
}

type TaskStatusResponse struct {
	ID          uuid.UUID `json:"id"`
	StatusName  string    `json:"status_name"`
	ColorName   string    `json:"color_name"`
	ColorHex    string    `json:"color_hex"`
	StatusOrder int       `json:"status_order"`
	IsDone      bool      `json:"is_done"`
	IsQAStatus  bool      `json:"is_qa_status"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewResponse(model TaskStatus) TaskStatusResponse {
	return TaskStatusResponse{
		ID:          model.ID,
		StatusName:  model.StatusName,
		ColorName:   model.ColorName,
		ColorHex:    model.ColorHex,
		StatusOrder: model.StatusOrder,
		IsDone:      model.IsDone,
		IsQAStatus:  model.IsQAStatus,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func NewResponses(models []TaskStatus) []TaskStatusResponse {
	result := make([]TaskStatusResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}
