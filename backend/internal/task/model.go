package task

import (
	"time"

	"devtracker/backend/internal/project"
	"devtracker/backend/internal/sprint"
	"devtracker/backend/internal/status"
	"devtracker/backend/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID              uuid.UUID         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProjectID       uuid.UUID         `gorm:"type:uuid;not null;index"`
	Project         project.Project   `gorm:"foreignKey:ProjectID"`
	SprintID        uuid.UUID         `gorm:"type:uuid;not null;index"`
	Sprint          sprint.Sprint     `gorm:"foreignKey:SprintID"`
	DeveloperID     uuid.UUID         `gorm:"type:uuid;not null;index"`
	Developer       user.User         `gorm:"foreignKey:DeveloperID"`
	StatusID        uuid.UUID         `gorm:"type:uuid;not null;index"`
	Status          status.TaskStatus `gorm:"foreignKey:StatusID"`
	TicketNumber    string            `gorm:"size:100"`
	TaskTitle       string            `gorm:"size:255;not null"`
	TaskDescription string            `gorm:"type:text"`
	TaskType        string            `gorm:"size:100"`
	Priority        string            `gorm:"size:50;not null;default:medium"`
	EstimatedPoint  *float64          `gorm:"type:numeric(10,2)"`
	ActualPoint     *float64          `gorm:"type:numeric(10,2)"`
	StartDate       *time.Time        `gorm:"type:date"`
	DueDate         *time.Time        `gorm:"type:date"`
	CompletedDate   *time.Time
	QACheckedDate   *time.Time
	CreatedBy       *uuid.UUID `gorm:"type:uuid"`
	UpdatedBy       *uuid.UUID `gorm:"type:uuid"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type TaskHistory struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TaskID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	OldStatusID *uuid.UUID `gorm:"type:uuid"`
	NewStatusID uuid.UUID  `gorm:"type:uuid;not null"`
	ChangedBy   uuid.UUID  `gorm:"type:uuid;not null"`
	ChangedAt   time.Time  `gorm:"not null;default:now()"`
	Note        string     `gorm:"type:text"`
}

func (TaskHistory) TableName() string {
	return "task_histories"
}

func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	return nil
}

func (h *TaskHistory) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}

	if h.ChangedAt.IsZero() {
		h.ChangedAt = time.Now().UTC()
	}

	return nil
}
