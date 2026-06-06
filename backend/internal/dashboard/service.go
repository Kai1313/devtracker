package dashboard

import (
	"context"
	"errors"
	"strings"

	"devtracker/backend/internal/sprint"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repository Repository
	sprints    sprint.Repository
}

func NewService(repository Repository, sprints sprint.Repository) *Service {
	return &Service{
		repository: repository,
		sprints:    sprints,
	}
}

func (s *Service) Summary(ctx context.Context, query SummaryQuery) (*SummaryResponse, error) {
	var sprintID *uuid.UUID

	if strings.TrimSpace(query.SprintID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(query.SprintID))
		if err != nil {
			return nil, apperrors.BadRequest("sprint_id must be a valid UUID")
		}

		if _, err := s.sprints.FindByID(ctx, parsed); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.BadRequest("sprint does not exist")
			}

			return nil, err
		}

		sprintID = &parsed
	}

	return s.repository.Summary(ctx, sprintID)
}
