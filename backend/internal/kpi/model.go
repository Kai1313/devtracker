package kpi

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPISnapshot struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	DeveloperID         uuid.UUID `gorm:"type:uuid;not null;index"`
	SprintID            uuid.UUID `gorm:"type:uuid;not null;index"`
	TotalAssigned       int64     `gorm:"not null;default:0"`
	TotalDone           int64     `gorm:"not null;default:0"`
	TotalReadyToCheck   int64     `gorm:"not null;default:0"`
	TotalQAChecked      int64     `gorm:"column:total_qa_checked;not null;default:0"`
	DelayedTaskCount    int64     `gorm:"not null;default:0"`
	CompletionRate      float64   `gorm:"not null;default:0"`
	QAPassRate          float64   `gorm:"not null;default:0"`
	TotalEstimatedPoint float64   `gorm:"not null;default:0"`
	TotalActualPoint    float64   `gorm:"not null;default:0"`
	CalculatedAt        time.Time `gorm:"not null;default:now()"`
}

func (KPISnapshot) TableName() string {
	return "kpi_snapshots"
}

func (s *KPISnapshot) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
