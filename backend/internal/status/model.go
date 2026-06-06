package status

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	StatusName  string    `gorm:"size:100;uniqueIndex;not null"`
	ColorName   string    `gorm:"size:30;not null"`
	ColorHex    string    `gorm:"size:7;not null"`
	StatusOrder int       `gorm:"not null;default:0"`
	IsDone      bool      `gorm:"not null;default:false"`
	IsQAStatus  bool      `gorm:"column:is_qa_status;not null;default:false"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (TaskStatus) TableName() string {
	return "task_statuses"
}

func (s *TaskStatus) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
