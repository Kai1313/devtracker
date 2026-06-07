package project

import (
	"context"
	"errors"
	"strings"
	"time"

	appquery "devtracker/backend/internal/query"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

type Service struct {
	repository Repository
}

var projectSortFields = map[string]string{
	"project_code": "project_code",
	"project_name": "project_name",
	"client_name":  "client_name",
	"status":       "status",
	"start_date":   "start_date",
	"end_date":     "end_date",
	"created_at":   "created_at",
	"updated_at":   "updated_at",
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, filter ListProjectsQuery) ([]ProjectResponse, map[string]any, error) {
	filter.Page = appquery.NormalizePage(filter.Page)
	filter.Limit = appquery.NormalizeLimit(filter.Limit)
	filter.Search = strings.TrimSpace(filter.Search)

	sort, err := appquery.NormalizeSort(filter.SortBy, filter.SortOrder, projectSortFields, appquery.Sort{By: "created_at", Order: appquery.Descending})
	if err != nil {
		return nil, nil, err
	}
	filter.SortBy = sort.By
	filter.SortOrder = sort.Order

	projects, total, err := s.repository.List(ctx, filter)
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

	return NewResponses(projects), meta, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*ProjectResponse, error) {
	project, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("project not found")
		}

		return nil, err
	}

	response := NewResponse(*project)
	return &response, nil
}

func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*ProjectResponse, error) {
	projectCode := normalizeText(req.ProjectCode)
	if projectCode == "" {
		return nil, apperrors.BadRequest("project_code cannot be empty")
	}

	projectName := normalizeText(req.ProjectName)
	if projectName == "" {
		return nil, apperrors.BadRequest("project_name cannot be empty")
	}

	if err := s.ensureProjectCodeAvailable(ctx, projectCode, uuid.Nil); err != nil {
		return nil, err
	}

	startDate, endDate, err := parseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	project := &Project{
		ID:          uuid.New(),
		ProjectCode: projectCode,
		ProjectName: projectName,
		ClientName:  normalizeText(req.ClientName),
		Status:      normalizeStatus(req.Status),
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := s.repository.Create(ctx, project); err != nil {
		return nil, err
	}

	response := NewResponse(*project)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	project, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("project not found")
		}

		return nil, err
	}

	if req.ProjectCode != nil {
		projectCode := normalizeText(*req.ProjectCode)
		if projectCode == "" {
			return nil, apperrors.BadRequest("project_code cannot be empty")
		}

		if err := s.ensureProjectCodeAvailable(ctx, projectCode, project.ID); err != nil {
			return nil, err
		}

		project.ProjectCode = projectCode
	}

	if req.ProjectName != nil {
		projectName := normalizeText(*req.ProjectName)
		if projectName == "" {
			return nil, apperrors.BadRequest("project_name cannot be empty")
		}

		project.ProjectName = projectName
	}

	if req.ClientName != nil {
		project.ClientName = normalizeText(*req.ClientName)
	}

	if req.Status != nil {
		project.Status = normalizeStatus(*req.Status)
	}

	if req.StartDate != nil {
		startDate, err := parseOptionalDate(*req.StartDate, "start_date")
		if err != nil {
			return nil, err
		}

		project.StartDate = startDate
	}

	if req.EndDate != nil {
		endDate, err := parseOptionalDate(*req.EndDate, "end_date")
		if err != nil {
			return nil, err
		}

		project.EndDate = endDate
	}

	if err := validateDateRange(project.StartDate, project.EndDate); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, project); err != nil {
		return nil, err
	}

	response := NewResponse(*project)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) ensureProjectCodeAvailable(ctx context.Context, code string, currentID uuid.UUID) error {
	existing, err := s.repository.FindByCodeIncludingDeleted(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	if currentID == uuid.Nil || existing.ID != currentID {
		return apperrors.Conflict("project_code is already registered")
	}

	return nil
}

func parseDateRange(startDateValue string, endDateValue string) (*time.Time, *time.Time, error) {
	startDate, err := parseOptionalDate(startDateValue, "start_date")
	if err != nil {
		return nil, nil, err
	}

	endDate, err := parseOptionalDate(endDateValue, "end_date")
	if err != nil {
		return nil, nil, err
	}

	if err := validateDateRange(startDate, endDate); err != nil {
		return nil, nil, err
	}

	return startDate, endDate, nil
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

func validateDateRange(startDate *time.Time, endDate *time.Time) error {
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return apperrors.BadRequest("start_date cannot be after end_date")
	}

	return nil
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func normalizeStatus(value string) string {
	status := strings.TrimSpace(value)
	if status == "" {
		return "active"
	}

	return status
}
