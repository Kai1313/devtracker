package status

import (
	"context"
	"errors"
	"regexp"
	"strings"

	appquery "devtracker/backend/internal/query"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var hexColorPattern = regexp.MustCompile(`^#([A-Fa-f0-9]{3}|[A-Fa-f0-9]{6})$`)

type Service struct {
	repository Repository
}

var taskStatusSortFields = map[string]string{
	"status_name":  "status_name",
	"color_name":   "color_name",
	"color_hex":    "color_hex",
	"status_order": "status_order",
	"is_done":      "is_done",
	"is_qa_status": "is_qa_status",
	"is_active":    "is_active",
	"created_at":   "created_at",
	"updated_at":   "updated_at",
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, filter ListTaskStatusesQuery) ([]TaskStatusResponse, map[string]any, error) {
	filter.Page = appquery.NormalizePage(filter.Page)
	filter.Limit = appquery.NormalizeLimit(filter.Limit)
	filter.Search = strings.TrimSpace(filter.Search)

	sort, err := appquery.NormalizeSort(filter.SortBy, filter.SortOrder, taskStatusSortFields, appquery.Sort{By: "status_order", Order: appquery.Ascending})
	if err != nil {
		return nil, nil, err
	}
	filter.SortBy = sort.By
	filter.SortOrder = sort.Order

	statuses, total, err := s.repository.List(ctx, filter)
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

	return NewResponses(statuses), meta, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*TaskStatusResponse, error) {
	taskStatus, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("task status not found")
		}

		return nil, err
	}

	response := NewResponse(*taskStatus)
	return &response, nil
}

func (s *Service) Create(ctx context.Context, req CreateTaskStatusRequest) (*TaskStatusResponse, error) {
	statusName := normalizeText(req.StatusName)
	if statusName == "" {
		return nil, apperrors.BadRequest("status_name cannot be empty")
	}

	if err := s.ensureStatusNameAvailable(ctx, statusName, uuid.Nil); err != nil {
		return nil, err
	}

	colorName := normalizeText(req.ColorName)
	if colorName == "" {
		return nil, apperrors.BadRequest("color_name cannot be empty")
	}

	colorHex, err := normalizeColorHex(req.ColorHex)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	taskStatus := &TaskStatus{
		ID:          uuid.New(),
		StatusName:  statusName,
		ColorName:   colorName,
		ColorHex:    colorHex,
		StatusOrder: req.StatusOrder,
		IsDone:      req.IsDone,
		IsQAStatus:  req.IsQAStatus,
		IsActive:    isActive,
	}

	if err := s.repository.Create(ctx, taskStatus); err != nil {
		return nil, err
	}

	response := NewResponse(*taskStatus)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateTaskStatusRequest) (*TaskStatusResponse, error) {
	taskStatus, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("task status not found")
		}

		return nil, err
	}

	if req.StatusName != nil {
		statusName := normalizeText(*req.StatusName)
		if statusName == "" {
			return nil, apperrors.BadRequest("status_name cannot be empty")
		}

		if err := s.ensureStatusNameAvailable(ctx, statusName, taskStatus.ID); err != nil {
			return nil, err
		}

		taskStatus.StatusName = statusName
	}

	if req.ColorName != nil {
		colorName := normalizeText(*req.ColorName)
		if colorName == "" {
			return nil, apperrors.BadRequest("color_name cannot be empty")
		}

		taskStatus.ColorName = colorName
	}

	if req.ColorHex != nil {
		colorHex, err := normalizeColorHex(*req.ColorHex)
		if err != nil {
			return nil, err
		}

		taskStatus.ColorHex = colorHex
	}

	if req.StatusOrder != nil {
		taskStatus.StatusOrder = *req.StatusOrder
	}

	if req.IsDone != nil {
		taskStatus.IsDone = *req.IsDone
	}

	if req.IsQAStatus != nil {
		taskStatus.IsQAStatus = *req.IsQAStatus
	}

	if req.IsActive != nil {
		taskStatus.IsActive = *req.IsActive
	}

	if err := s.repository.Update(ctx, taskStatus); err != nil {
		return nil, err
	}

	response := NewResponse(*taskStatus)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) ensureStatusNameAvailable(ctx context.Context, name string, currentID uuid.UUID) error {
	existing, err := s.repository.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	if currentID == uuid.Nil || existing.ID != currentID {
		return apperrors.Conflict("status_name is already registered")
	}

	return nil
}

func normalizeColorHex(value string) (string, error) {
	colorHex := strings.ToUpper(strings.TrimSpace(value))
	if colorHex == "" {
		return "", apperrors.BadRequest("color_hex cannot be empty")
	}

	if !hexColorPattern.MatchString(colorHex) {
		return "", apperrors.BadRequest("color_hex must use #RGB or #RRGGBB format")
	}

	return colorHex, nil
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}
