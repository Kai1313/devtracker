package sprint

import (
	"context"
	"errors"
	"strings"
	"time"

	"devtracker/backend/internal/project"
	appquery "devtracker/backend/internal/query"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	dateLayout     = "2006-01-02"
	StatusPlanning = "planning"
	StatusActive   = "active"
	StatusClosed   = "closed"
)

type Service struct {
	repository  Repository
	projects    project.Repository
	snapshotter KPISnapshotter
}

var sprintSortFields = map[string]string{
	"sprint_name": "sprint_name",
	"start_date":  "start_date",
	"end_date":    "end_date",
	"status":      "status",
	"created_at":  "created_at",
	"updated_at":  "updated_at",
}

type KPISnapshotter interface {
	GenerateSprintSnapshots(ctx context.Context, sprintID uuid.UUID) error
}

func NewService(repository Repository, projects project.Repository, snapshotters ...KPISnapshotter) *Service {
	service := &Service{
		repository: repository,
		projects:   projects,
	}

	if len(snapshotters) > 0 {
		service.snapshotter = snapshotters[0]
	}

	return service
}

func (s *Service) List(ctx context.Context, filter ListSprintsQuery) ([]SprintResponse, map[string]any, error) {
	filter.Page = appquery.NormalizePage(filter.Page)
	filter.Limit = appquery.NormalizeLimit(filter.Limit)

	sort, err := appquery.NormalizeSort(filter.SortBy, filter.SortOrder, sprintSortFields, appquery.Sort{By: "start_date", Order: appquery.Descending})
	if err != nil {
		return nil, nil, err
	}
	filter.SortBy = sort.By
	filter.SortOrder = sort.Order

	if filter.ProjectID != "" {
		projectID, err := uuid.Parse(strings.TrimSpace(filter.ProjectID))
		if err != nil {
			return nil, nil, apperrors.BadRequest("project_id must be a valid UUID")
		}

		filter.ProjectID = projectID.String()
	}

	status, err := normalizeOptionalStatus(filter.Status)
	if err != nil {
		return nil, nil, err
	}
	filter.Status = status

	sprints, total, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]any{
		"page":       filter.Page,
		"limit":      filter.Limit,
		"total":      total,
		"sort_by":    filter.SortBy,
		"sort_order": filter.SortOrder,
	}

	return NewResponses(sprints), meta, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*SprintResponse, error) {
	sprint, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("sprint not found")
		}

		return nil, err
	}

	response := NewResponse(*sprint)
	return &response, nil
}

func (s *Service) Create(ctx context.Context, req CreateSprintRequest) (*SprintResponse, error) {
	projectID, linkedProject, err := s.resolveProject(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}

	sprintName := normalizeText(req.SprintName)
	if sprintName == "" {
		return nil, apperrors.BadRequest("sprint_name cannot be empty")
	}

	startDate, endDate, err := parseRequiredDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	status, err := normalizeOptionalStatus(req.Status)
	if err != nil {
		return nil, err
	}
	if status == "" {
		status = StatusPlanning
	}

	sprint := &Sprint{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Project:    *linkedProject,
		SprintName: sprintName,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     status,
	}

	if err := s.repository.Create(ctx, sprint); err != nil {
		return nil, err
	}

	response := NewResponse(*sprint)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateSprintRequest) (*SprintResponse, error) {
	current, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("sprint not found")
		}

		return nil, err
	}

	if req.ProjectID != nil {
		projectID, linkedProject, err := s.resolveProject(ctx, *req.ProjectID)
		if err != nil {
			return nil, err
		}

		current.ProjectID = projectID
		current.Project = *linkedProject
	}

	if req.SprintName != nil {
		sprintName := normalizeText(*req.SprintName)
		if sprintName == "" {
			return nil, apperrors.BadRequest("sprint_name cannot be empty")
		}

		current.SprintName = sprintName
	}

	if req.StartDate != nil {
		startDate, err := parseRequiredDate(*req.StartDate, "start_date")
		if err != nil {
			return nil, err
		}

		current.StartDate = startDate
	}

	if req.EndDate != nil {
		endDate, err := parseRequiredDate(*req.EndDate, "end_date")
		if err != nil {
			return nil, err
		}

		current.EndDate = endDate
	}

	wasClosed := current.Status == StatusClosed

	if req.Status != nil {
		status, err := normalizeRequiredStatus(*req.Status)
		if err != nil {
			return nil, err
		}

		current.Status = status
	}

	if err := validateDateRange(current.StartDate, current.EndDate); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, current); err != nil {
		return nil, err
	}

	if !wasClosed && current.Status == StatusClosed {
		if err := s.generateSnapshots(ctx, current.ID); err != nil {
			return nil, err
		}
	}

	response := NewResponse(*current)
	return &response, nil
}

func (s *Service) Close(ctx context.Context, id uuid.UUID) (*SprintResponse, error) {
	current, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("sprint not found")
		}

		return nil, err
	}

	wasClosed := current.Status == StatusClosed
	current.Status = StatusClosed

	if err := s.repository.Update(ctx, current); err != nil {
		return nil, err
	}

	if !wasClosed {
		if err := s.generateSnapshots(ctx, current.ID); err != nil {
			return nil, err
		}
	}

	response := NewResponse(*current)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) generateSnapshots(ctx context.Context, sprintID uuid.UUID) error {
	if s.snapshotter == nil {
		return nil
	}

	return s.snapshotter.GenerateSprintSnapshots(ctx, sprintID)
}

func (s *Service) resolveProject(ctx context.Context, value string) (uuid.UUID, *project.Project, error) {
	projectID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, nil, apperrors.BadRequest("project_id must be a valid UUID")
	}

	linkedProject, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil, apperrors.BadRequest("project does not exist")
		}

		return uuid.Nil, nil, err
	}

	return projectID, linkedProject, nil
}

func parseRequiredDateRange(startDateValue string, endDateValue string) (time.Time, time.Time, error) {
	startDate, err := parseRequiredDate(startDateValue, "start_date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err := parseRequiredDate(endDateValue, "end_date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if err := validateDateRange(startDate, endDate); err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

func parseRequiredDate(value string, field string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, apperrors.BadRequest(field + " is required")
	}

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, apperrors.BadRequest(field + " must use YYYY-MM-DD format")
	}

	return parsed, nil
}

func validateDateRange(startDate time.Time, endDate time.Time) error {
	if startDate.After(endDate) {
		return apperrors.BadRequest("start_date cannot be after end_date")
	}

	return nil
}

func normalizeOptionalStatus(value string) (string, error) {
	status := normalizeText(strings.ToLower(value))
	if status == "" {
		return "", nil
	}

	if !isValidStatus(status) {
		return "", apperrors.BadRequest("status must be planning, active, or closed")
	}

	return status, nil
}

func normalizeRequiredStatus(value string) (string, error) {
	status, err := normalizeOptionalStatus(value)
	if err != nil {
		return "", err
	}

	if status == "" {
		return "", apperrors.BadRequest("status is required")
	}

	return status, nil
}

func isValidStatus(status string) bool {
	switch status {
	case StatusPlanning, StatusActive, StatusClosed:
		return true
	default:
		return false
	}
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}
