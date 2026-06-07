package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title           string     `gorm:"size:150;not null"`
	Message         string     `gorm:"type:text;not null"`
	Type            string     `gorm:"size:100;not null;index"`
	ReferenceModule string     `gorm:"size:100;index"`
	ReferenceID     *uuid.UUID `gorm:"type:uuid;index"`
	IsRead          bool       `gorm:"not null;default:false;index"`
	ReadAt          *time.Time
	CreatedAt       time.Time
}

func (n *Notification) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	return nil
}
