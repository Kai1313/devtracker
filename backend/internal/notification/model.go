package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TaskID    *uuid.UUID `gorm:"type:uuid;index"`
	Type      string     `gorm:"size:100;not null;index"`
	Title     string     `gorm:"size:150;not null"`
	Message   string     `gorm:"type:text;not null"`
	IsRead    bool       `gorm:"not null;default:false;index"`
	ReadAt    *time.Time
	CreatedAt time.Time
}

func (n *Notification) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	return nil
}
