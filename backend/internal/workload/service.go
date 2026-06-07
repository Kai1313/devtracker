package workload

import (
	"context"
	"errors"
	"strings"

	"devtracker/backend/internal/project"
	"devtracker/backend/internal/sprint"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repository Repository
	sprints    sprint.Repository
	projects   project.Repository
}

func NewService(repository Repository, sprints sprint.Repository, projects project.Repository) *Service {
	return &Service{
		repository: repository,
		sprints:    sprints,
		projects:   projects,
	}
}

func (s *Service) DeveloperWorkload(ctx context.Context, query Query) ([]DeveloperWorkloadResponse, error) {
	filter, err := s.normalizeQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	return s.repository.DeveloperWorkload(ctx, filter)
}

func (s *Service) normalizeQuery(ctx context.Context, query Query) (filter, error) {
	var result filter

	if strings.TrimSpace(query.SprintID) != "" {
		sprintID, err := uuid.Parse(strings.TrimSpace(query.SprintID))
		if err != nil {
			return result, apperrors.BadRequest("sprint must be a valid UUID")
		}

		if _, err := s.sprints.FindByID(ctx, sprintID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return result, apperrors.BadRequest("sprint does not exist")
			}

			return result, err
		}

		result.SprintID = &sprintID
	}

	if strings.TrimSpace(query.ProjectID) != "" {
		projectID, err := uuid.Parse(strings.TrimSpace(query.ProjectID))
		if err != nil {
			return result, apperrors.BadRequest("project must be a valid UUID")
		}

		if _, err := s.projects.FindByID(ctx, projectID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return result, apperrors.BadRequest("project does not exist")
			}

			return result, err
		}

		result.ProjectID = &projectID
	}

	return result, nil
}

func classify(totalPoints float64) string {
	switch {
	case totalPoints < 5:
		return ClassificationLow
	case totalPoints <= 13:
		return ClassificationNormal
	case totalPoints <= 20:
		return ClassificationHigh
	default:
		return ClassificationOverloaded
	}
}
