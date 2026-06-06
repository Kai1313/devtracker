package project

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProjectCode string     `gorm:"size:50;uniqueIndex;not null"`
	ProjectName string     `gorm:"size:150;not null"`
	ClientName  string     `gorm:"size:150"`
	Status      string     `gorm:"size:50;not null;default:active"`
	StartDate   *time.Time `gorm:"type:date"`
	EndDate     *time.Time `gorm:"type:date"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (p *Project) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}
