package task

import (
	"context"
	"errors"
	"strings"
	"time"

	"devtracker/backend/internal/project"
	appquery "devtracker/backend/internal/query"
	"devtracker/backend/internal/sprint"
	"devtracker/backend/internal/status"
	"devtracker/backend/internal/user"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	dateLayout     = "2006-01-02"
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

type Service struct {
	repository Repository
	users      user.Repository
	projects   project.Repository
	sprints    sprint.Repository
	statuses   status.Repository
}

var taskSortFields = map[string]string{
	"ticket_number":   "ticket_number",
	"task_title":      "task_title",
	"priority":        "priority",
	"estimated_point": "estimated_point",
	"actual_point":    "actual_point",
	"start_date":      "start_date",
	"due_date":        "due_date",
	"completed_date":  "completed_date",
	"qa_checked_date": "qa_checked_date",
	"created_at":      "created_at",
	"updated_at":      "updated_at",
}

type AccessScope struct {
	UserID                 uuid.UUID
	CanManageTasks         bool
	CanViewAssignedTasks   bool
	CanViewReadyToCheck    bool
	CanUpdateOwnTaskStatus bool
	CanUpdateQAStatus      bool
}

func NewService(
	repository Repository,
	users user.Repository,
	projects project.Repository,
	sprints sprint.Repository,
	statuses status.Repository,
) *Service {
	return &Service{
		repository: repository,
		users:      users,
		projects:   projects,
		sprints:    sprints,
		statuses:   statuses,
	}
}

func (s *Service) List(ctx context.Context, filter ListTasksQuery) ([]TaskResponse, map[string]any, error) {
	normalized, err := prepareListFilter(filter)
	if err != nil {
		return nil, nil, err
	}

	return s.list(ctx, normalized, ListAccessFilter{})
}

func (s *Service) ListWithAccess(ctx context.Context, filter ListTasksQuery, scope AccessScope) ([]TaskResponse, map[string]any, error) {
	normalized, err := prepareListFilter(filter)
	if err != nil {
		return nil, nil, err
	}

	if scope.CanManageTasks {
		return s.list(ctx, normalized, ListAccessFilter{})
	}

	access, err := s.listAccessFilter(ctx, normalized, scope)
	if err != nil {
		return nil, nil, err
	}

	return s.list(ctx, normalized, access)
}

func (s *Service) list(ctx context.Context, filter ListTasksQuery, access ListAccessFilter) ([]TaskResponse, map[string]any, error) {
	var (
		tasks []Task
		total int64
		err   error
	)

	if access.IsZero() {
		tasks, total, err = s.repository.List(ctx, filter)
	} else {
		tasks, total, err = s.repository.ListWithAccess(ctx, filter, access)
	}
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

	return NewResponses(tasks), meta, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*TaskResponse, error) {
	current, err := s.findTask(ctx, id)
	if err != nil {
		return nil, err
	}

	response := NewResponse(*current)
	return &response, nil
}

func (s *Service) GetWithAccess(ctx context.Context, id uuid.UUID, scope AccessScope) (*TaskResponse, error) {
	current, err := s.findTask(ctx, id)
	if err != nil {
		return nil, err
	}

	if !canViewTask(current, scope) {
		return nil, apperrors.Forbidden("insufficient permissions for this task")
	}

	response := NewResponse(*current)
	return &response, nil
}

func (s *Service) findTask(ctx context.Context, id uuid.UUID) (*Task, error) {
	current, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("task not found")
		}

		return nil, err
	}

	return current, nil
}

func (s *Service) ListHistories(ctx context.Context, taskID uuid.UUID) ([]TaskHistoryResponse, error) {
	if _, err := s.Get(ctx, taskID); err != nil {
		return nil, err
	}

	histories, err := s.repository.ListHistories(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return NewHistoryResponses(histories), nil
}

func (s *Service) ListHistoriesWithAccess(ctx context.Context, taskID uuid.UUID, scope AccessScope) ([]TaskHistoryResponse, error) {
	if _, err := s.GetWithAccess(ctx, taskID, scope); err != nil {
		return nil, err
	}

	histories, err := s.repository.ListHistories(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return NewHistoryResponses(histories), nil
}

func (s *Service) Create(ctx context.Context, req CreateTaskRequest, actorID uuid.UUID) (*TaskResponse, error) {
	refs, err := s.resolveReferences(ctx, referenceInput{
		developerID: req.DeveloperID,
		projectID:   req.ProjectID,
		sprintID:    req.SprintID,
		statusID:    req.StatusID,
	})
	if err != nil {
		return nil, err
	}

	taskTitle := normalizeText(req.TaskTitle)
	if taskTitle == "" {
		return nil, apperrors.BadRequest("task_title cannot be empty")
	}

	priority, err := normalizeRequiredPriority(req.Priority)
	if err != nil {
		return nil, err
	}

	startDate, err := parseOptionalDate(req.StartDate, "start_date")
	if err != nil {
		return nil, err
	}

	dueDate, err := parseOptionalDate(req.DueDate, "due_date")
	if err != nil {
		return nil, err
	}

	if err := validateDateRange(startDate, dueDate); err != nil {
		return nil, err
	}

	completedDate, err := parseOptionalTimestamp(req.CompletedDate, "completed_date")
	if err != nil {
		return nil, err
	}

	qaCheckedDate, err := parseOptionalTimestamp(req.QACheckedDate, "qa_checked_date")
	if err != nil {
		return nil, err
	}

	applyStatusChangeDates(refs.status, &completedDate, &qaCheckedDate)

	task := &Task{
		ID:              uuid.New(),
		ProjectID:       refs.project.ID,
		Project:         *refs.project,
		SprintID:        refs.sprint.ID,
		Sprint:          *refs.sprint,
		DeveloperID:     refs.developer.ID,
		Developer:       *refs.developer,
		StatusID:        refs.status.ID,
		Status:          *refs.status,
		TicketNumber:    normalizeText(req.TicketNumber),
		TaskTitle:       taskTitle,
		TaskDescription: strings.TrimSpace(req.TaskDescription),
		Priority:        priority,
		EstimatedPoint:  req.EstimatedPoint,
		ActualPoint:     req.ActualPoint,
		StartDate:       startDate,
		DueDate:         dueDate,
		CompletedDate:   completedDate,
		QACheckedDate:   qaCheckedDate,
		CreatedBy:       &actorID,
		UpdatedBy:       &actorID,
	}

	history := newStatusHistory(task.ID, nil, task.StatusID, actorID, req.Note)
	if err := s.repository.Create(ctx, task, history); err != nil {
		return nil, err
	}

	response := NewResponse(*task)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateTaskRequest, actorID uuid.UUID) (*TaskResponse, error) {
	current, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("task not found")
		}

		return nil, err
	}

	oldStatusID := current.StatusID

	if req.DeveloperID != nil {
		developer, err := s.resolveDeveloper(ctx, *req.DeveloperID)
		if err != nil {
			return nil, err
		}

		current.DeveloperID = developer.ID
		current.Developer = *developer
	}

	if req.ProjectID != nil {
		linkedProject, err := s.resolveProject(ctx, *req.ProjectID)
		if err != nil {
			return nil, err
		}

		current.ProjectID = linkedProject.ID
		current.Project = *linkedProject
	}

	if req.SprintID != nil {
		linkedSprint, err := s.resolveSprint(ctx, *req.SprintID)
		if err != nil {
			return nil, err
		}

		current.SprintID = linkedSprint.ID
		current.Sprint = *linkedSprint
	}

	if current.Sprint.ProjectID != current.ProjectID {
		return nil, apperrors.BadRequest("sprint_id must belong to project_id")
	}

	if req.StatusID != nil {
		taskStatus, err := s.resolveStatus(ctx, *req.StatusID)
		if err != nil {
			return nil, err
		}

		current.StatusID = taskStatus.ID
		current.Status = *taskStatus
	}

	if req.TicketNumber != nil {
		current.TicketNumber = normalizeText(*req.TicketNumber)
	}

	if req.TaskTitle != nil {
		taskTitle := normalizeText(*req.TaskTitle)
		if taskTitle == "" {
			return nil, apperrors.BadRequest("task_title cannot be empty")
		}

		current.TaskTitle = taskTitle
	}

	if req.TaskDescription != nil {
		current.TaskDescription = strings.TrimSpace(*req.TaskDescription)
	}

	if req.Priority != nil {
		priority, err := normalizeRequiredPriority(*req.Priority)
		if err != nil {
			return nil, err
		}

		current.Priority = priority
	}

	if req.EstimatedPoint != nil {
		current.EstimatedPoint = req.EstimatedPoint
	}

	if req.ActualPoint != nil {
		current.ActualPoint = req.ActualPoint
	}

	if req.StartDate != nil {
		startDate, err := parseOptionalDate(*req.StartDate, "start_date")
		if err != nil {
			return nil, err
		}

		current.StartDate = startDate
	}

	if req.DueDate != nil {
		dueDate, err := parseOptionalDate(*req.DueDate, "due_date")
		if err != nil {
			return nil, err
		}

		current.DueDate = dueDate
	}

	if err := validateDateRange(current.StartDate, current.DueDate); err != nil {
		return nil, err
	}

	if req.CompletedDate != nil {
		completedDate, err := parseOptionalTimestamp(*req.CompletedDate, "completed_date")
		if err != nil {
			return nil, err
		}

		current.CompletedDate = completedDate
	}

	if req.QACheckedDate != nil {
		qaCheckedDate, err := parseOptionalTimestamp(*req.QACheckedDate, "qa_checked_date")
		if err != nil {
			return nil, err
		}

		current.QACheckedDate = qaCheckedDate
	}

	current.UpdatedBy = &actorID

	var history *TaskHistory
	if oldStatusID != current.StatusID {
		applyStatusChangeDates(&current.Status, &current.CompletedDate, &current.QACheckedDate)

		note := ""
		if req.Note != nil {
			note = *req.Note
		}

		history = newStatusHistory(current.ID, &oldStatusID, current.StatusID, actorID, note)
	}

	if err := s.repository.Update(ctx, current, history); err != nil {
		return nil, err
	}

	response := NewResponse(*current)
	return &response, nil
}

func (s *Service) UpdateWithAccess(ctx context.Context, id uuid.UUID, req UpdateTaskRequest, actorID uuid.UUID, scope AccessScope) (*TaskResponse, error) {
	if scope.CanManageTasks {
		return s.Update(ctx, id, req, actorID)
	}

	current, err := s.findTask(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isStatusOnlyUpdate(req) {
		return nil, apperrors.Forbidden("only task status can be updated")
	}

	if req.StatusID == nil || strings.TrimSpace(*req.StatusID) == "" {
		return nil, apperrors.BadRequest("status_id is required")
	}

	if scope.CanUpdateOwnTaskStatus && current.DeveloperID == scope.UserID {
		return s.Update(ctx, id, req, actorID)
	}

	if scope.CanUpdateQAStatus {
		if !isReadyToCheckStatus(&current.Status) && !isQAStatus(&current.Status) {
			return nil, apperrors.Forbidden("QA can only update tasks in QA statuses")
		}

		nextStatus, err := s.resolveStatus(ctx, *req.StatusID)
		if err != nil {
			return nil, err
		}

		if !isQAStatus(nextStatus) {
			return nil, apperrors.Forbidden("QA can only update QA-related statuses")
		}

		return s.Update(ctx, id, req, actorID)
	}

	return nil, apperrors.Forbidden("insufficient permissions for this task")
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) resolveReferences(ctx context.Context, input referenceInput) (*resolvedReferences, error) {
	developer, err := s.resolveDeveloper(ctx, input.developerID)
	if err != nil {
		return nil, err
	}

	linkedProject, err := s.resolveProject(ctx, input.projectID)
	if err != nil {
		return nil, err
	}

	linkedSprint, err := s.resolveSprint(ctx, input.sprintID)
	if err != nil {
		return nil, err
	}

	if linkedSprint.ProjectID != linkedProject.ID {
		return nil, apperrors.BadRequest("sprint_id must belong to project_id")
	}

	taskStatus, err := s.resolveStatus(ctx, input.statusID)
	if err != nil {
		return nil, err
	}

	return &resolvedReferences{
		developer: developer,
		project:   linkedProject,
		sprint:    linkedSprint,
		status:    taskStatus,
	}, nil
}

func (s *Service) resolveDeveloper(ctx context.Context, value string) (*user.User, error) {
	developerID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil, apperrors.BadRequest("developer_id must be a valid UUID")
	}

	developer, err := s.users.FindByID(ctx, developerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("developer does not exist")
		}

		return nil, err
	}

	if !developer.IsActive {
		return nil, apperrors.BadRequest("developer must be active")
	}

	if developer.Role.Name != "developer" {
		return nil, apperrors.BadRequest("developer_id must reference a developer user")
	}

	return developer, nil
}

func (s *Service) resolveProject(ctx context.Context, value string) (*project.Project, error) {
	projectID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil, apperrors.BadRequest("project_id must be a valid UUID")
	}

	linkedProject, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("project does not exist")
		}

		return nil, err
	}

	return linkedProject, nil
}

func (s *Service) resolveSprint(ctx context.Context, value string) (*sprint.Sprint, error) {
	sprintID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil, apperrors.BadRequest("sprint_id must be a valid UUID")
	}

	linkedSprint, err := s.sprints.FindByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("sprint does not exist")
		}

		return nil, err
	}

	return linkedSprint, nil
}

func (s *Service) resolveStatus(ctx context.Context, value string) (*status.TaskStatus, error) {
	statusID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil, apperrors.BadRequest("status_id must be a valid UUID")
	}

	taskStatus, err := s.statuses.FindByID(ctx, statusID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("status does not exist")
		}

		return nil, err
	}

	if !taskStatus.IsActive {
		return nil, apperrors.BadRequest("status must be active")
	}

	return taskStatus, nil
}

func newStatusHistory(taskID uuid.UUID, oldStatusID *uuid.UUID, newStatusID uuid.UUID, actorID uuid.UUID, note string) *TaskHistory {
	return &TaskHistory{
		ID:          uuid.New(),
		TaskID:      taskID,
		OldStatusID: oldStatusID,
		NewStatusID: newStatusID,
		ChangedBy:   actorID,
		ChangedAt:   time.Now().UTC(),
		Note:        normalizeText(note),
	}
}

func prepareListFilter(filter ListTasksQuery) (ListTasksQuery, error) {
	filter.Page = appquery.NormalizePage(filter.Page)
	filter.Limit = appquery.NormalizeLimit(filter.Limit)
	filter.Search = strings.TrimSpace(filter.Search)

	sort, err := appquery.NormalizeSort(filter.SortBy, filter.SortOrder, taskSortFields, appquery.Sort{By: "created_at", Order: appquery.Descending})
	if err != nil {
		return ListTasksQuery{}, err
	}
	filter.SortBy = sort.By
	filter.SortOrder = sort.Order

	if err := normalizeFilterIDs(&filter); err != nil {
		return ListTasksQuery{}, err
	}

	return filter, nil
}

func (s *Service) listAccessFilter(ctx context.Context, filter ListTasksQuery, scope AccessScope) (ListAccessFilter, error) {
	if scope.UserID == uuid.Nil {
		return ListAccessFilter{}, apperrors.Forbidden("insufficient permissions")
	}

	access := ListAccessFilter{}

	if scope.CanViewAssignedTasks {
		access.DeveloperID = scope.UserID
	}

	if scope.CanViewReadyToCheck {
		readyStatus, err := s.statuses.FindByName(ctx, "Ready to Check")
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ListAccessFilter{}, apperrors.NotFound("ready to check status not found")
			}

			return ListAccessFilter{}, err
		}

		access.ReadyToCheckStatusID = readyStatus.ID
	}

	if access.IsZero() {
		return ListAccessFilter{}, apperrors.Forbidden("insufficient permissions")
	}

	if filter.DeveloperID != "" && access.DeveloperID != uuid.Nil && filter.DeveloperID != access.DeveloperID.String() && access.ReadyToCheckStatusID == uuid.Nil {
		return ListAccessFilter{}, apperrors.Forbidden("developers can only view assigned tasks")
	}

	if filter.StatusID != "" && access.ReadyToCheckStatusID != uuid.Nil && filter.StatusID != access.ReadyToCheckStatusID.String() && access.DeveloperID == uuid.Nil {
		return ListAccessFilter{}, apperrors.Forbidden("QA can only view tasks ready to check")
	}

	return access, nil
}

func normalizeFilterIDs(filter *ListTasksQuery) error {
	normalized, err := normalizeOptionalUUID(filter.DeveloperID, "developer_id")
	if err != nil {
		return err
	}
	filter.DeveloperID = normalized

	normalized, err = normalizeOptionalUUID(filter.ProjectID, "project_id")
	if err != nil {
		return err
	}
	filter.ProjectID = normalized

	normalized, err = normalizeOptionalUUID(filter.SprintID, "sprint_id")
	if err != nil {
		return err
	}
	filter.SprintID = normalized

	normalized, err = normalizeOptionalUUID(filter.StatusID, "status_id")
	if err != nil {
		return err
	}
	filter.StatusID = normalized

	return nil
}

func normalizeOptionalUUID(value string, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return "", apperrors.BadRequest(field + " must be a valid UUID")
	}

	return parsed.String(), nil
}

func normalizeRequiredPriority(value string) (string, error) {
	priority := strings.ToLower(normalizeText(value))
	if priority == "" {
		return "", apperrors.BadRequest("priority is required")
	}

	if !isValidPriority(priority) {
		return "", apperrors.BadRequest("priority must be low, medium, high, or urgent")
	}

	return priority, nil
}

func isValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
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

func parseOptionalTimestamp(value string, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		normalized := parsed.UTC()
		return &normalized, nil
	}

	dateValue, err := time.Parse(dateLayout, value)
	if err == nil {
		return &dateValue, nil
	}

	return nil, apperrors.BadRequest(field + " must use RFC3339 or YYYY-MM-DD format")
}

func validateDateRange(startDate *time.Time, dueDate *time.Time) error {
	if startDate != nil && dueDate != nil && startDate.After(*dueDate) {
		return apperrors.BadRequest("start_date cannot be after due_date")
	}

	return nil
}

func applyStatusChangeDates(taskStatus *status.TaskStatus, completedDate **time.Time, qaCheckedDate **time.Time) {
	now := time.Now().UTC()

	statusName := strings.ToLower(strings.TrimSpace(taskStatus.StatusName))

	if (taskStatus.IsDone || statusName == "done") && *completedDate == nil {
		*completedDate = &now
	}

	if statusName == "checked by qa" && *qaCheckedDate == nil {
		*qaCheckedDate = &now
	}
}

func canViewTask(task *Task, scope AccessScope) bool {
	if scope.CanManageTasks {
		return true
	}

	if scope.CanViewAssignedTasks && task.DeveloperID == scope.UserID {
		return true
	}

	return scope.CanViewReadyToCheck && isReadyToCheckStatus(&task.Status)
}

func isStatusOnlyUpdate(req UpdateTaskRequest) bool {
	return req.DeveloperID == nil &&
		req.ProjectID == nil &&
		req.SprintID == nil &&
		req.TicketNumber == nil &&
		req.TaskTitle == nil &&
		req.TaskDescription == nil &&
		req.Priority == nil &&
		req.EstimatedPoint == nil &&
		req.ActualPoint == nil &&
		req.StartDate == nil &&
		req.DueDate == nil &&
		req.CompletedDate == nil &&
		req.QACheckedDate == nil
}

func isReadyToCheckStatus(taskStatus *status.TaskStatus) bool {
	if taskStatus == nil {
		return false
	}

	return normalizeStatusName(taskStatus.StatusName) == "ready_to_check"
}

func isQAStatus(taskStatus *status.TaskStatus) bool {
	if taskStatus == nil {
		return false
	}

	statusName := normalizeStatusName(taskStatus.StatusName)
	return taskStatus.IsQAStatus || statusName == "ready_to_check" || statusName == "checked_by_qa"
}

func normalizeStatusName(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	return normalized
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

type referenceInput struct {
	developerID string
	projectID   string
	sprintID    string
	statusID    string
}

type resolvedReferences struct {
	developer *user.User
	project   *project.Project
	sprint    *sprint.Sprint
	status    *status.TaskStatus
}
