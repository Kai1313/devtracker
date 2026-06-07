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
	Action    string
	StartDate string
	EndDate   string
}

type listFilter struct {
	Page      int
	Limit     int
	UserID    *uuid.UUID
	Module    string
	Modules   []string
	Action    string
	StartDate *time.Time
	EndDate   *time.Time
}

type ListScope struct {
	CanViewAll     bool
	AllowedModules []string
}

type RecordInput struct {
	UserID    *uuid.UUID
	Module    string
	Action    string
	EntityID  *uuid.UUID
	OldValue  any
	NewValue  any
	IPAddress string
	UserAgent string
}

type AuditLogResponse struct {
	ID        uuid.UUID        `json:"id"`
	UserID    *uuid.UUID       `json:"user_id,omitempty"`
	Module    string           `json:"module"`
	Action    string           `json:"action"`
	EntityID  *uuid.UUID       `json:"entity_id,omitempty"`
	OldValue  *json.RawMessage `json:"old_value,omitempty"`
	NewValue  *json.RawMessage `json:"new_value,omitempty"`
	IPAddress string           `json:"ip_address,omitempty"`
	UserAgent string           `json:"user_agent,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

func NewResponse(model AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:        model.ID,
		UserID:    model.UserID,
		Module:    model.Module,
		Action:    model.Action,
		EntityID:  model.EntityID,
		OldValue:  rawJSON(model.OldValue),
		NewValue:  rawJSON(model.NewValue),
		IPAddress: model.IPAddress,
		UserAgent: model.UserAgent,
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

func rawJSON(value JSONValue) *json.RawMessage {
	if len(value) == 0 {
		return nil
	}

	raw := json.RawMessage(value)
	return &raw
}
