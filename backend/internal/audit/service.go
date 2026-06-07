package audit

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

var projectManagerAuditModules = []string{"projects", "sprints", "tasks"}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, input RecordInput) error {
	module := normalize(input.Module)
	if module == "" {
		return apperrors.BadRequest("module is required")
	}

	action := normalize(input.Action)
	if action == "" {
		return apperrors.BadRequest("action is required")
	}

	oldValue, err := marshalValue(input.OldValue)
	if err != nil {
		return err
	}

	newValue, err := marshalValue(input.NewValue)
	if err != nil {
		return err
	}

	return s.repository.Create(ctx, &AuditLog{
		ID:        uuid.New(),
		UserID:    input.UserID,
		Module:    module,
		Action:    action,
		EntityID:  input.EntityID,
		OldValue:  oldValue,
		NewValue:  newValue,
		IPAddress: strings.TrimSpace(input.IPAddress),
		UserAgent: strings.TrimSpace(input.UserAgent),
	})
}

func (s *Service) List(ctx context.Context, query ListQuery) ([]AuditLogResponse, map[string]any, error) {
	filter, err := normalizeListQuery(query)
	if err != nil {
		return nil, nil, err
	}

	logs, total, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]any{
		"page":  filter.Page,
		"limit": filter.Limit,
		"total": total,
	}

	return NewResponses(logs), meta, nil
}

func (s *Service) ListWithScope(ctx context.Context, query ListQuery, scope ListScope) ([]AuditLogResponse, map[string]any, error) {
	filter, err := normalizeListQuery(query)
	if err != nil {
		return nil, nil, err
	}

	if !scope.CanViewAll {
		allowedModules := normalizeModules(scope.AllowedModules)
		if len(allowedModules) == 0 {
			return nil, nil, apperrors.Forbidden("insufficient permissions")
		}

		if filter.Module != "" && !slices.Contains(allowedModules, filter.Module) {
			return nil, nil, apperrors.Forbidden("insufficient permissions for this audit module")
		}

		filter.Modules = allowedModules
	}

	logs, total, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]any{
		"page":  filter.Page,
		"limit": filter.Limit,
		"total": total,
	}

	return NewResponses(logs), meta, nil
}

func normalizeListQuery(query ListQuery) (listFilter, error) {
	filter := listFilter{
		Page:   normalizePage(query.Page),
		Limit:  normalizeLimit(query.Limit),
		Module: normalize(query.Module),
		Action: normalize(query.Action),
	}

	if strings.TrimSpace(query.UserID) != "" {
		userID, err := uuid.Parse(strings.TrimSpace(query.UserID))
		if err != nil {
			return filter, apperrors.BadRequest("user_id must be a valid UUID")
		}

		filter.UserID = &userID
	}

	startDate, err := parseOptionalDate(query.StartDate, "start_date")
	if err != nil {
		return filter, err
	}
	filter.StartDate = startDate

	endDate, err := parseOptionalDate(query.EndDate, "end_date")
	if err != nil {
		return filter, err
	}
	if endDate != nil {
		nextDay := endDate.AddDate(0, 0, 1)
		filter.EndDate = &nextDay
	}

	if filter.StartDate != nil && filter.EndDate != nil && filter.StartDate.After(*filter.EndDate) {
		return filter, apperrors.BadRequest("start_date cannot be after end_date")
	}

	return filter, nil
}

func marshalValue(value any) (JSONValue, error) {
	if value == nil {
		return nil, nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return JSONValue(payload), nil
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

func normalize(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	return normalized
}

func normalizeModules(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		normalized := normalize(value)
		if normalized == "" {
			continue
		}

		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}

	return page
}

func normalizeLimit(limit int) int {
	if limit < 1 {
		return 20
	}

	if limit > 100 {
		return 100
	}

	return limit
}
