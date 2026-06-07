package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	Module    string     `gorm:"size:100;not null;index"`
	Action    string     `gorm:"size:50;not null;index"`
	OldValue  *string    `gorm:"type:text"`
	NewValue  *string    `gorm:"type:text"`
	IPAddress string     `gorm:"size:100"`
	CreatedAt time.Time
}

func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	return nil
}
