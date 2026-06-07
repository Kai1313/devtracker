package query

import (
	"slices"
	"strings"

	apperrors "devtracker/backend/pkg/errors"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
	Ascending    = "asc"
	Descending   = "desc"
)

type Sort struct {
	By    string
	Order string
}

func NormalizePage(page int) int {
	if page < 1 {
		return DefaultPage
	}

	return page
}

func NormalizeLimit(limit int) int {
	if limit < 1 {
		return DefaultLimit
	}

	if limit > MaxLimit {
		return MaxLimit
	}

	return limit
}

func NormalizeSort(sortBy, sortOrder string, allowed map[string]string, fallback Sort) (Sort, error) {
	normalizedBy := normalizeName(sortBy)
	if normalizedBy == "" {
		normalizedBy = fallback.By
	}

	if _, ok := allowed[normalizedBy]; !ok {
		return Sort{}, apperrors.BadRequest("sort_by must be one of: " + strings.Join(allowedKeys(allowed), ", "))
	}

	normalizedOrder := strings.ToLower(strings.TrimSpace(sortOrder))
	if normalizedOrder == "" {
		normalizedOrder = fallback.Order
	}

	if normalizedOrder != Ascending && normalizedOrder != Descending {
		return Sort{}, apperrors.BadRequest("sort_order must be asc or desc")
	}

	return Sort{By: normalizedBy, Order: normalizedOrder}, nil
}

func OrderClause(sort Sort, allowed map[string]string) string {
	column := allowed[sort.By]
	if column == "" {
		return ""
	}

	return column + " " + strings.ToUpper(sort.Order)
}

func normalizeName(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	return normalized
}

func allowedKeys(allowed map[string]string) []string {
	keys := make([]string, 0, len(allowed))
	for key := range allowed {
		keys = append(keys, key)
	}

	slices.Sort(keys)
	return keys
}
