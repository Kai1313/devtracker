package sprint

import (
	"time"

	"devtracker/backend/internal/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Sprint struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProjectID  uuid.UUID       `gorm:"type:uuid;not null;index"`
	Project    project.Project `gorm:"foreignKey:ProjectID"`
	SprintName string          `gorm:"size:150;not null"`
	StartDate  time.Time       `gorm:"type:date;not null"`
	EndDate    time.Time       `gorm:"type:date;not null"`
	Status     string          `gorm:"size:50;not null;default:active"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (s *Sprint) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
