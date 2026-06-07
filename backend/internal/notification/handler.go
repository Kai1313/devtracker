package notification

import (
	"devtracker/backend/internal/audit"
	"devtracker/backend/internal/httpx"
	appmiddleware "devtracker/backend/internal/middleware"
	"devtracker/backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
	audit   *audit.Service
}

func NewHandler(service *Service, auditService *audit.Service) *Handler {
	return &Handler{service: service, audit: auditService}
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	result, meta, err := h.service.List(c.UserContext(), ListQuery{
		Page:       c.QueryInt("page", 1),
		Limit:      c.QueryInt("limit", 20),
		UserID:     userID,
		IncludeAll: isAdmin(c),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
	})
	if err != nil {
		return err
	}

	return response.WithMeta(c, "notifications retrieved", result, meta)
}

func (h *Handler) UnreadCount(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	result, err := h.service.UnreadCount(c.UserContext(), userID, isAdmin(c))
	if err != nil {
		return err
	}

	return response.OK(c, "notification unread count retrieved", result)
}

func (h *Handler) MarkRead(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, changed, err := h.service.MarkRead(c.UserContext(), id, userID, isAdmin(c))
	if err != nil {
		return err
	}

	if changed {
		if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
			UserID:   &userID,
			Module:   "notifications",
			Action:   "notification_read",
			EntityID: &id,
			OldValue: map[string]any{
				"is_read": false,
			},
			NewValue: result.Notification,
		}); err != nil {
			return err
		}
	}

	return response.OK(c, "notification marked as read", result)
}

func (h *Handler) MarkAllRead(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	result, err := h.service.MarkAllRead(c.UserContext(), userID, isAdmin(c))
	if err != nil {
		return err
	}

	if result.ReadCount > 0 {
		if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
			UserID: &userID,
			Module: "notifications",
			Action: "notification_read_all",
			NewValue: map[string]any{
				"read_count":   result.ReadCount,
				"unread_count": result.UnreadCount,
			},
		}); err != nil {
			return err
		}
	}

	return response.OK(c, "notifications marked as read", result)
}

func isAdmin(c *fiber.Ctx) bool {
	return appmiddleware.HasRole(c, "admin")
}
