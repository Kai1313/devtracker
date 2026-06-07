package audit

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	Module    string     `gorm:"size:100;not null;index"`
	Action    string     `gorm:"size:50;not null;index"`
	EntityID  *uuid.UUID `gorm:"type:uuid;index"`
	OldValue  JSONValue  `gorm:"type:jsonb"`
	NewValue  JSONValue  `gorm:"type:jsonb"`
	IPAddress string     `gorm:"size:100"`
	UserAgent string     `gorm:"type:text"`
	CreatedAt time.Time
}

type JSONValue []byte

func (j JSONValue) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}

	if !json.Valid(j) {
		return nil, fmt.Errorf("invalid JSON value")
	}

	return string(j), nil
}

func (j *JSONValue) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch typed := value.(type) {
	case []byte:
		*j = append((*j)[:0], typed...)
	case string:
		*j = append((*j)[:0], typed...)
	default:
		return fmt.Errorf("unsupported JSON value type %T", value)
	}

	return nil
}

func (j JSONValue) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}

	return j, nil
}

func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	return nil
}
