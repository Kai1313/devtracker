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
	sprintID, err := s.resolveSprintID(ctx, query.SprintID)
	if err != nil {
		return nil, err
	}

	return s.repository.DeveloperKPI(ctx, sprintID)
}

func (s *Service) Projects(ctx context.Context, query Query) ([]ProjectKPIResponse, error) {
	sprintID, err := s.resolveSprintID(ctx, query.SprintID)
	if err != nil {
		return nil, err
	}

	return s.repository.ProjectKPI(ctx, sprintID)
}

func (s *Service) resolveSprintID(ctx context.Context, value string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, apperrors.BadRequest("sprint_id must be a valid UUID")
	}

	if _, err := s.sprints.FindByID(ctx, parsed); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("sprint does not exist")
		}

		return nil, err
	}

	return &parsed, nil
}
