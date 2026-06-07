package kpi

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPISnapshot struct {
	ID                         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	SprintID                   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_kpi_snapshots_sprint_developer"`
	DeveloperID                uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_kpi_snapshots_sprint_developer"`
	TotalAssignedTasks         int64     `gorm:"not null;default:0"`
	TotalDoneTasks             int64     `gorm:"not null;default:0"`
	TotalReadyToCheckTasks     int64     `gorm:"not null;default:0"`
	TotalCheckedByQATasks      int64     `gorm:"not null;default:0"`
	DelayedTasks               int64     `gorm:"not null;default:0"`
	CompletionRate             float64   `gorm:"not null;default:0"`
	TotalEstimatedPoints       float64   `gorm:"not null;default:0"`
	TotalActualPoints          float64   `gorm:"not null;default:0"`
	AverageCompletionTimeHours float64   `gorm:"not null;default:0"`
	GeneratedAt                time.Time `gorm:"not null;default:now()"`
	CreatedAt                  time.Time
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
