package kpi

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

func (s *Service) Developers(ctx context.Context, query Query) ([]DeveloperKPIResponse, error) {
	sprintID, closed, err := s.resolveSprint(ctx, query.SprintID)
	if err != nil {
		return nil, err
	}

	if sprintID != nil && closed {
		return s.repository.DeveloperSnapshotKPI(ctx, *sprintID)
	}

	return s.repository.DeveloperKPI(ctx, sprintID)
}

func (s *Service) Projects(ctx context.Context, query Query) ([]ProjectKPIResponse, error) {
	sprintID, closed, err := s.resolveSprint(ctx, query.SprintID)
	if err != nil {
		return nil, err
	}

	if sprintID != nil && closed {
		return s.repository.ProjectSnapshotKPI(ctx, *sprintID)
	}

	return s.repository.ProjectKPI(ctx, sprintID)
}

func (s *Service) GenerateSprintSnapshots(ctx context.Context, sprintID uuid.UUID) error {
	return s.repository.GenerateSprintSnapshots(ctx, sprintID)
}

func (s *Service) resolveSprint(ctx context.Context, value string) (*uuid.UUID, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, false, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, false, apperrors.BadRequest("sprint_id must be a valid UUID")
	}

	current, err := s.sprints.FindByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, apperrors.BadRequest("sprint does not exist")
		}

		return nil, false, err
	}

	return &parsed, strings.ToLower(strings.TrimSpace(current.Status)) == sprint.StatusClosed, nil
}
