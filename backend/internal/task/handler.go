package task

import (
	"strings"

	"devtracker/backend/internal/audit"
	"devtracker/backend/internal/httpx"
	"devtracker/backend/internal/notification"
	apperrors "devtracker/backend/pkg/errors"
	"devtracker/backend/pkg/response"
	appvalidator "devtracker/backend/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service       *Service
	audit         *audit.Service
	notifications *notification.Service
}

func NewHandler(service *Service, auditService *audit.Service, notificationService *notification.Service) *Handler {
	return &Handler{service: service, audit: auditService, notifications: notificationService}
}

func (h *Handler) List(c *fiber.Ctx) error {
	query := ListTasksQuery{
		Page:        c.QueryInt("page", 1),
		Limit:       c.QueryInt("limit", 20),
		DeveloperID: c.Query("developer_id"),
		ProjectID:   c.Query("project_id"),
		SprintID:    c.Query("sprint_id"),
		StatusID:    c.Query("status_id"),
		Search:      c.Query("search"),
	}

	result, meta, err := h.service.List(c.UserContext(), query)
	if err != nil {
		return err
	}

	return response.WithMeta(c, "tasks retrieved", result, meta)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "task retrieved", result)
}

func (h *Handler) ListHistories(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := h.service.ListHistories(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "task histories retrieved", result)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), req, actorID)
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "tasks",
		Action:   "create",
		NewValue: result,
	}); err != nil {
		return err
	}

	if err := notification.CreateNotification(c.UserContext(), h.notifications, taskAssignedNotification(result)); err != nil {
		return err
	}

	return response.Created(c, "task created", result)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	oldValue, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), id, req, actorID)
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "tasks",
		Action:   "update",
		OldValue: oldValue,
		NewValue: result,
	}); err != nil {
		return err
	}

	if oldValue.StatusID != result.StatusID {
		if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
			UserID:   &actorID,
			Module:   "tasks",
			Action:   "status_change",
			OldValue: taskStatusAuditValue(oldValue),
			NewValue: taskStatusAuditValue(result),
		}); err != nil {
			return err
		}
	}

	if oldValue.DeveloperID != result.DeveloperID {
		if err := notification.CreateNotification(c.UserContext(), h.notifications, taskAssignedNotification(result)); err != nil {
			return err
		}
	}

	if oldValue.StatusID != result.StatusID {
		input, ok := statusChangeNotification(result)
		if ok {
			if err := notification.CreateNotification(c.UserContext(), h.notifications, input); err != nil {
				return err
			}
		}
	}

	return response.OK(c, "task updated", result)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	oldValue, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "tasks",
		Action:   "delete",
		OldValue: oldValue,
	}); err != nil {
		return err
	}

	return response.OK(c, "task deleted", nil)
}

func taskStatusAuditValue(task *TaskResponse) map[string]any {
	return map[string]any{
		"task_id":     task.ID,
		"status_id":   task.StatusID,
		"status_name": task.Status.StatusName,
	}
}

func taskAssignedNotification(task *TaskResponse) notification.CreateInput {
	taskID := task.ID
	return notification.CreateInput{
		UserID:  task.DeveloperID,
		TaskID:  &taskID,
		Type:    notification.TypeTaskAssigned,
		Title:   "Task assigned",
		Message: "Task assigned: " + task.TaskTitle,
	}
}

func statusChangeNotification(task *TaskResponse) (notification.CreateInput, bool) {
	taskID := task.ID
	input := notification.CreateInput{
		UserID: task.DeveloperID,
		TaskID: &taskID,
	}

	switch strings.ToLower(strings.TrimSpace(task.Status.StatusName)) {
	case "ready to check":
		input.Type = notification.TypeTaskReadyToCheck
		input.Title = "Task ready to check"
		input.Message = "Task moved to Ready to Check: " + task.TaskTitle
	case "checked by qa":
		input.Type = notification.TypeTaskCheckedByQA
		input.Title = "Task checked by QA"
		input.Message = "Task moved to Checked by QA: " + task.TaskTitle
	case "done":
		input.Type = notification.TypeTaskDone
		input.Title = "Task done"
		input.Message = "Task moved to Done: " + task.TaskTitle
	default:
		return notification.CreateInput{}, false
	}

	return input, true
}
