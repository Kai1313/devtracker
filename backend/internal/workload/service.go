package workload

import (
	"context"
	"errors"
	"strings"
	"time"

	"devtracker/backend/internal/project"
	"devtracker/backend/internal/sprint"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

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
	return s.DeveloperWorkloadWithScope(ctx, query, AccessScope{IsAdmin: true})
}

func (s *Service) DeveloperWorkloadWithScope(ctx context.Context, query Query, scope AccessScope) ([]DeveloperWorkloadResponse, error) {
	filter, err := s.normalizeQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	if err := applyAccessScope(&filter, scope); err != nil {
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

	if strings.TrimSpace(query.DeveloperID) != "" {
		developerID, err := uuid.Parse(strings.TrimSpace(query.DeveloperID))
		if err != nil {
			return result, apperrors.BadRequest("developer_id must be a valid UUID")
		}

		result.DeveloperID = &developerID
	}

	if strings.TrimSpace(query.StatusID) != "" {
		statusID, err := uuid.Parse(strings.TrimSpace(query.StatusID))
		if err != nil {
			return result, apperrors.BadRequest("status_id must be a valid UUID")
		}

		result.StatusID = &statusID
	}

	startDate, err := parseOptionalDate(query.StartDate, "start_date")
	if err != nil {
		return result, err
	}
	result.StartDate = startDate

	endDate, err := parseOptionalDate(query.EndDate, "end_date")
	if err != nil {
		return result, err
	}
	result.EndDate = endDate

	if result.StartDate != nil && result.EndDate != nil && result.StartDate.After(*result.EndDate) {
		return result, apperrors.BadRequest("start_date cannot be after end_date")
	}

	return result, nil
}

func applyAccessScope(filter *filter, scope AccessScope) error {
	if scope.IsAdmin || scope.IsManager || scope.IsManagement {
		return nil
	}

	if scope.IsDeveloper && !scope.IsQA {
		if scope.UserID == uuid.Nil {
			return apperrors.Forbidden("insufficient permissions")
		}

		if filter.DeveloperID != nil && *filter.DeveloperID != scope.UserID {
			return apperrors.Forbidden("developers can only view their own workload")
		}

		filter.DeveloperID = &scope.UserID
		return nil
	}

	if scope.IsQA {
		filter.QAOnly = true
		return nil
	}

	return apperrors.Forbidden("insufficient permissions")
}

func parseOptionalDate(value string, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil, apperrors.BadRequest(field + " must use YYYY-MM-DD format")
	}

	return &parsed, nil
}

func classify(activeTasks int64) string {
	switch {
	case activeTasks <= 3:
		return ClassificationLow
	case activeTasks <= 7:
		return ClassificationNormal
	case activeTasks <= 10:
		return ClassificationHigh
	default:
		return ClassificationOverloaded
	}
}
