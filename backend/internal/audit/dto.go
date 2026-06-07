package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ListQuery struct {
	Page      int
	Limit     int
	UserID    string
	Module    string
	StartDate string
	EndDate   string
}

type listFilter struct {
	Page      int
	Limit     int
	UserID    *uuid.UUID
	Module    string
	StartDate *time.Time
	EndDate   *time.Time
}

type RecordInput struct {
	UserID    *uuid.UUID
	Module    string
	Action    string
	OldValue  any
	NewValue  any
	IPAddress string
}

type AuditLogResponse struct {
	ID        uuid.UUID        `json:"id"`
	UserID    *uuid.UUID       `json:"user_id,omitempty"`
	Module    string           `json:"module"`
	Action    string           `json:"action"`
	OldValue  *json.RawMessage `json:"old_value,omitempty"`
	NewValue  *json.RawMessage `json:"new_value,omitempty"`
	IPAddress string           `json:"ip_address,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

func NewResponse(model AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:        model.ID,
		UserID:    model.UserID,
		Module:    model.Module,
		Action:    model.Action,
		OldValue:  rawJSON(model.OldValue),
		NewValue:  rawJSON(model.NewValue),
		IPAddress: model.IPAddress,
		CreatedAt: model.CreatedAt,
	}
}

func NewResponses(models []AuditLog) []AuditLogResponse {
	result := make([]AuditLogResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}

func rawJSON(value *string) *json.RawMessage {
	if value == nil || *value == "" {
		return nil
	}

	raw := json.RawMessage(*value)
	return &raw
}
