package task

import (
	"strings"
	"time"

	"devtracker/backend/internal/audit"
	"devtracker/backend/internal/httpx"
	appmiddleware "devtracker/backend/internal/middleware"
	"devtracker/backend/internal/notification"
	apperrors "devtracker/backend/pkg/errors"
	"devtracker/backend/pkg/response"
	appvalidator "devtracker/backend/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	scope, err := taskAccessScope(c)
	if err != nil {
		return err
	}

	query := ListTasksQuery{
		Page:        c.QueryInt("page", 1),
		Limit:       c.QueryInt("limit", 20),
		DeveloperID: c.Query("developer_id"),
		ProjectID:   c.Query("project_id"),
		SprintID:    c.Query("sprint_id"),
		StatusID:    c.Query("status_id"),
		Search:      c.Query("search"),
	}

	result, meta, err := h.service.ListWithAccess(c.UserContext(), query, scope)
	if err != nil {
		return err
	}

	return response.WithMeta(c, "tasks retrieved", result, meta)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	scope, err := taskAccessScope(c)
	if err != nil {
		return err
	}

	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := h.service.GetWithAccess(c.UserContext(), id, scope)
	if err != nil {
		return err
	}

	return response.OK(c, "task retrieved", result)
}

func (h *Handler) ListHistories(c *fiber.Ctx) error {
	scope, err := taskAccessScope(c)
	if err != nil {
		return err
	}

	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := h.service.ListHistoriesWithAccess(c.UserContext(), id, scope)
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
		EntityID: &result.ID,
		NewValue: result,
	}); err != nil {
		return err
	}

	if err := notification.CreateNotification(c.UserContext(), h.notifications, taskAssignedNotification(result)); err != nil {
		return err
	}

	if isTaskOverdue(result) {
		if err := notification.CreateNotification(c.UserContext(), h.notifications, taskOverdueNotification(result)); err != nil {
			return err
		}
	}

	return response.Created(c, "task created", result)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}
	scope := taskAccessScopeForUser(c, actorID)

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

	result, err := h.service.UpdateWithAccess(c.UserContext(), id, req, actorID, scope)
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "tasks",
		Action:   "update",
		EntityID: &result.ID,
		OldValue: oldValue,
		NewValue: result,
	}); err != nil {
		return err
	}

	if oldValue.StatusID != result.StatusID {
		if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
			UserID:   &actorID,
			Module:   "tasks",
			Action:   "task_status_change",
			EntityID: &result.ID,
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

	if !isTaskOverdue(oldValue) && isTaskOverdue(result) {
		if err := notification.CreateNotification(c.UserContext(), h.notifications, taskOverdueNotification(result)); err != nil {
			return err
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
		EntityID: &id,
		OldValue: oldValue,
	}); err != nil {
		return err
	}

	return response.OK(c, "task deleted", nil)
}

func taskAccessScope(c *fiber.Ctx) (AccessScope, error) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return AccessScope{}, err
	}

	return taskAccessScopeForUser(c, userID), nil
}

func taskAccessScopeForUser(c *fiber.Ctx, userID uuid.UUID) AccessScope {
	return AccessScope{
		UserID:                 userID,
		CanManageTasks:         appmiddleware.HasPermission(c, "manage_tasks"),
		CanViewAssignedTasks:   appmiddleware.HasPermission(c, "view_assigned_tasks"),
		CanViewReadyToCheck:    appmiddleware.HasPermission(c, "view_ready_to_check_tasks"),
		CanUpdateOwnTaskStatus: appmiddleware.HasPermission(c, "update_own_task_status"),
		CanUpdateQAStatus:      appmiddleware.HasPermission(c, "update_qa_status"),
	}
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
		UserID:          task.DeveloperID,
		Type:            notification.TypeTaskAssigned,
		Title:           "Task assigned",
		Message:         "Task assigned: " + task.TaskTitle,
		ReferenceModule: notification.ReferenceModuleTasks,
		ReferenceID:     &taskID,
	}
}

func statusChangeNotification(task *TaskResponse) (notification.CreateInput, bool) {
	taskID := task.ID
	input := notification.CreateInput{
		UserID:          task.DeveloperID,
		ReferenceModule: notification.ReferenceModuleTasks,
		ReferenceID:     &taskID,
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

func taskOverdueNotification(task *TaskResponse) notification.CreateInput {
	taskID := task.ID
	return notification.CreateInput{
		UserID:          task.DeveloperID,
		Type:            notification.TypeTaskOverdue,
		Title:           "Task overdue",
		Message:         "Task overdue: " + task.TaskTitle,
		ReferenceModule: notification.ReferenceModuleTasks,
		ReferenceID:     &taskID,
	}
}

func isTaskOverdue(task *TaskResponse) bool {
	if task == nil || task.DueDate == nil || task.CompletedDate != nil || task.Status.IsDone {
		return false
	}

	dueDate, err := time.Parse(dateLayout, *task.DueDate)
	if err != nil {
		return false
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	return dueDate.Before(today)
}
