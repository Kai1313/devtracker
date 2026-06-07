package task

import (
	"context"
	"strings"
	"testing"

	"devtracker/backend/internal/project"
	"devtracker/backend/internal/sprint"
	"devtracker/backend/internal/status"
	"devtracker/backend/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestTaskAssignmentCRUDAndStatusHistory(t *testing.T) {
	projectID := uuid.New()
	sprintID := uuid.New()
	developerID := uuid.New()
	actorID := uuid.New()
	todoStatusID := uuid.New()
	checkedByQAStatusID := uuid.New()
	doneStatusID := uuid.New()

	projectModel := &project.Project{ID: projectID, ProjectCode: "DEV", ProjectName: "Dev Tracker"}
	sprintModel := &sprint.Sprint{ID: sprintID, ProjectID: projectID, Project: *projectModel, SprintName: "Sprint 1"}
	developerRole := user.Role{ID: uuid.New(), Name: "developer"}
	developer := &user.User{
		ID:       developerID,
		RoleID:   developerRole.ID,
		Role:     developerRole,
		Name:     "Dev User",
		Email:    "dev@example.com",
		IsActive: true,
	}
	todoStatus := &status.TaskStatus{ID: todoStatusID, StatusName: "Todo", ColorName: "gray", ColorHex: "#6B7280", IsActive: true}
	checkedByQAStatus := &status.TaskStatus{ID: checkedByQAStatusID, StatusName: "Checked by QA", ColorName: "orange", ColorHex: "#F97316", IsQAStatus: true, IsActive: true}
	doneStatus := &status.TaskStatus{ID: doneStatusID, StatusName: "Done", ColorName: "green", ColorHex: "#22C55E", IsDone: true, IsActive: true}

	taskRepository := newFakeTaskRepository()
	service := NewService(
		taskRepository,
		&fakeUserRepository{users: map[uuid.UUID]*user.User{developerID: developer}},
		&fakeProjectRepository{projects: map[uuid.UUID]*project.Project{projectID: projectModel}},
		&fakeSprintRepository{sprints: map[uuid.UUID]*sprint.Sprint{sprintID: sprintModel}},
		&fakeStatusRepository{statuses: map[uuid.UUID]*status.TaskStatus{
			todoStatusID:        todoStatus,
			checkedByQAStatusID: checkedByQAStatus,
			doneStatusID:        doneStatus,
		}},
	)

	created, err := service.Create(context.Background(), CreateTaskRequest{
		DeveloperID:  developerID.String(),
		ProjectID:    projectID.String(),
		SprintID:     sprintID.String(),
		StatusID:     todoStatusID.String(),
		TicketNumber: "DEV-1",
		TaskTitle:    " Build API ",
		Priority:     PriorityMedium,
		StartDate:    "2026-01-01",
		DueDate:      "2026-01-05",
		Note:         "created",
	}, actorID)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if created.TaskTitle != "Build API" {
		t.Fatalf("expected trimmed title, got %q", created.TaskTitle)
	}
	if len(taskRepository.histories) != 1 {
		t.Fatalf("expected create history, got %d", len(taskRepository.histories))
	}
	if taskRepository.histories[0].OldStatusID != nil {
		t.Fatal("expected create history old status to be nil")
	}

	list, meta, err := service.List(context.Background(), ListTasksQuery{
		Page:        1,
		Limit:       10,
		DeveloperID: developerID.String(),
		ProjectID:   projectID.String(),
		SprintID:    sprintID.String(),
		StatusID:    todoStatusID.String(),
		Search:      "DEV-1",
	})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(list) != 1 || meta["total"].(int64) != 1 {
		t.Fatalf("expected one listed task, got len=%d meta=%v", len(list), meta)
	}

	checkedID := checkedByQAStatusID.String()
	checkedNote := "qa checked"
	checked, err := service.Update(context.Background(), created.ID, UpdateTaskRequest{
		StatusID: &checkedID,
		Note:     &checkedNote,
	}, actorID)
	if err != nil {
		t.Fatalf("update task to checked by QA: %v", err)
	}
	if checked.QACheckedDate == nil {
		t.Fatal("expected qa_checked_date to be set")
	}
	if len(taskRepository.histories) != 2 {
		t.Fatalf("expected second history after QA status change, got %d", len(taskRepository.histories))
	}
	if taskRepository.histories[1].Note != checkedNote {
		t.Fatalf("expected history note %q, got %q", checkedNote, taskRepository.histories[1].Note)
	}

	doneID := doneStatusID.String()
	done, err := service.Update(context.Background(), created.ID, UpdateTaskRequest{
		StatusID: &doneID,
	}, actorID)
	if err != nil {
		t.Fatalf("update task to done: %v", err)
	}
	if done.CompletedDate == nil {
		t.Fatal("expected completed_date to be set")
	}
	if len(taskRepository.histories) != 3 {
		t.Fatalf("expected third history after done status change, got %d", len(taskRepository.histories))
	}

	histories, err := service.ListHistories(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("list histories: %v", err)
	}
	if len(histories) != 3 {
		t.Fatalf("expected three histories, got %d", len(histories))
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if !taskRepository.deleted[created.ID] {
		t.Fatal("expected repository delete to be called")
	}
}

func TestTaskAccessScopeRestrictsDeveloperAndQA(t *testing.T) {
	projectID := uuid.New()
	sprintID := uuid.New()
	developerID := uuid.New()
	otherDeveloperID := uuid.New()
	qaID := uuid.New()
	readyStatusID := uuid.New()
	todoStatusID := uuid.New()
	checkedByQAStatusID := uuid.New()
	doneStatusID := uuid.New()

	projectModel := &project.Project{ID: projectID, ProjectCode: "DEV", ProjectName: "Dev Tracker"}
	sprintModel := &sprint.Sprint{ID: sprintID, ProjectID: projectID, Project: *projectModel, SprintName: "Sprint 1"}
	developerRole := user.Role{ID: uuid.New(), Name: "developer"}
	developer := &user.User{
		ID:       developerID,
		RoleID:   developerRole.ID,
		Role:     developerRole,
		Name:     "Dev User",
		Email:    "dev@example.com",
		IsActive: true,
	}
	otherDeveloper := &user.User{
		ID:       otherDeveloperID,
		RoleID:   developerRole.ID,
		Role:     developerRole,
		Name:     "Other Dev",
		Email:    "other@example.com",
		IsActive: true,
	}

	readyStatus := &status.TaskStatus{ID: readyStatusID, StatusName: "Ready to Check", ColorName: "blue", ColorHex: "#3B82F6", IsQAStatus: true, IsActive: true}
	todoStatus := &status.TaskStatus{ID: todoStatusID, StatusName: "Todo", ColorName: "gray", ColorHex: "#6B7280", IsActive: true}
	checkedByQAStatus := &status.TaskStatus{ID: checkedByQAStatusID, StatusName: "Checked by QA", ColorName: "orange", ColorHex: "#F97316", IsQAStatus: true, IsActive: true}
	doneStatus := &status.TaskStatus{ID: doneStatusID, StatusName: "Done", ColorName: "green", ColorHex: "#22C55E", IsDone: true, IsActive: true}

	taskRepository := newFakeTaskRepository()
	readyTask := &Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Project:     *projectModel,
		SprintID:    sprintID,
		Sprint:      *sprintModel,
		DeveloperID: developerID,
		Developer:   *developer,
		StatusID:    readyStatusID,
		Status:      *readyStatus,
		TaskTitle:   "Ready task",
		Priority:    PriorityMedium,
	}
	otherTask := &Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Project:     *projectModel,
		SprintID:    sprintID,
		Sprint:      *sprintModel,
		DeveloperID: otherDeveloperID,
		Developer:   *otherDeveloper,
		StatusID:    todoStatusID,
		Status:      *todoStatus,
		TaskTitle:   "Other task",
		Priority:    PriorityMedium,
	}
	taskRepository.tasks[readyTask.ID] = readyTask
	taskRepository.tasks[otherTask.ID] = otherTask

	service := NewService(
		taskRepository,
		&fakeUserRepository{users: map[uuid.UUID]*user.User{
			developerID:      developer,
			otherDeveloperID: otherDeveloper,
		}},
		&fakeProjectRepository{projects: map[uuid.UUID]*project.Project{projectID: projectModel}},
		&fakeSprintRepository{sprints: map[uuid.UUID]*sprint.Sprint{sprintID: sprintModel}},
		&fakeStatusRepository{statuses: map[uuid.UUID]*status.TaskStatus{
			readyStatusID:       readyStatus,
			todoStatusID:        todoStatus,
			checkedByQAStatusID: checkedByQAStatus,
			doneStatusID:        doneStatus,
		}},
	)

	developerScope := AccessScope{UserID: developerID, CanViewAssignedTasks: true, CanUpdateOwnTaskStatus: true}
	developerTasks, meta, err := service.ListWithAccess(context.Background(), ListTasksQuery{Page: 1, Limit: 10}, developerScope)
	if err != nil {
		t.Fatalf("developer scoped list: %v", err)
	}
	if len(developerTasks) != 1 || meta["total"].(int64) != 1 || developerTasks[0].ID != readyTask.ID {
		t.Fatalf("expected developer to see only assigned task, got tasks=%v meta=%v", developerTasks, meta)
	}

	if _, err := service.GetWithAccess(context.Background(), otherTask.ID, developerScope); err == nil {
		t.Fatal("expected developer to be forbidden from viewing another developer task")
	}

	renamed := "Renamed"
	if _, err := service.UpdateWithAccess(context.Background(), readyTask.ID, UpdateTaskRequest{TaskTitle: &renamed}, developerID, developerScope); err == nil {
		t.Fatal("expected developer non-status update to be forbidden")
	}

	qaScope := AccessScope{UserID: qaID, CanViewReadyToCheck: true, CanUpdateQAStatus: true}
	qaTasks, _, err := service.ListWithAccess(context.Background(), ListTasksQuery{Page: 1, Limit: 10}, qaScope)
	if err != nil {
		t.Fatalf("QA scoped list: %v", err)
	}
	if len(qaTasks) != 1 || qaTasks[0].ID != readyTask.ID {
		t.Fatalf("expected QA to see ready-to-check task only, got %v", qaTasks)
	}

	checkedID := checkedByQAStatusID.String()
	if _, err := service.UpdateWithAccess(context.Background(), readyTask.ID, UpdateTaskRequest{StatusID: &checkedID}, qaID, qaScope); err != nil {
		t.Fatalf("expected QA status update to be allowed: %v", err)
	}

	doneID := doneStatusID.String()
	if _, err := service.UpdateWithAccess(context.Background(), readyTask.ID, UpdateTaskRequest{StatusID: &doneID}, qaID, qaScope); err == nil {
		t.Fatal("expected QA update to non-QA status to be forbidden")
	}
}

type fakeTaskRepository struct {
	tasks     map[uuid.UUID]*Task
	histories []TaskHistory
	deleted   map[uuid.UUID]bool
}

func newFakeTaskRepository() *fakeTaskRepository {
	return &fakeTaskRepository{
		tasks:   map[uuid.UUID]*Task{},
		deleted: map[uuid.UUID]bool{},
	}
}

func (r *fakeTaskRepository) Create(_ context.Context, task *Task, history *TaskHistory) error {
	r.tasks[task.ID] = task
	if history != nil {
		r.histories = append(r.histories, *history)
	}
	return nil
}

func (r *fakeTaskRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted[id] = true
	return nil
}

func (r *fakeTaskRepository) FindByID(_ context.Context, id uuid.UUID) (*Task, error) {
	task, ok := r.tasks[id]
	if !ok || r.deleted[id] {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (r *fakeTaskRepository) List(_ context.Context, filter ListTasksQuery) ([]Task, int64, error) {
	return r.list(filter, ListAccessFilter{})
}

func (r *fakeTaskRepository) ListWithAccess(_ context.Context, filter ListTasksQuery, access ListAccessFilter) ([]Task, int64, error) {
	return r.list(filter, access)
}

func (r *fakeTaskRepository) list(filter ListTasksQuery, access ListAccessFilter) ([]Task, int64, error) {
	tasks := make([]Task, 0, len(r.tasks))
	for id, task := range r.tasks {
		if r.deleted[id] {
			continue
		}

		if !access.IsZero() && !fakeTaskMatchesAccess(task, access) {
			continue
		}

		if filter.DeveloperID != "" && task.DeveloperID.String() != filter.DeveloperID {
			continue
		}

		if filter.ProjectID != "" && task.ProjectID.String() != filter.ProjectID {
			continue
		}

		if filter.SprintID != "" && task.SprintID.String() != filter.SprintID {
			continue
		}

		if filter.StatusID != "" && task.StatusID.String() != filter.StatusID {
			continue
		}

		tasks = append(tasks, *task)
	}
	return tasks, int64(len(tasks)), nil
}

func fakeTaskMatchesAccess(task *Task, access ListAccessFilter) bool {
	if access.DeveloperID != uuid.Nil && task.DeveloperID == access.DeveloperID {
		return true
	}

	return access.ReadyToCheckStatusID != uuid.Nil && task.StatusID == access.ReadyToCheckStatusID
}

func (r *fakeTaskRepository) ListHistories(_ context.Context, taskID uuid.UUID) ([]TaskHistory, error) {
	histories := make([]TaskHistory, 0, len(r.histories))
	for _, history := range r.histories {
		if history.TaskID == taskID {
			histories = append(histories, history)
		}
	}
	return histories, nil
}

func (r *fakeTaskRepository) Update(_ context.Context, task *Task, history *TaskHistory) error {
	r.tasks[task.ID] = task
	if history != nil {
		r.histories = append(r.histories, *history)
	}
	return nil
}

type fakeUserRepository struct {
	users map[uuid.UUID]*user.User
}

func (r *fakeUserRepository) Count(context.Context) (int64, error) {
	return int64(len(r.users)), nil
}

func (r *fakeUserRepository) CountAll(context.Context) (int64, error) {
	return int64(len(r.users)), nil
}

func (r *fakeUserRepository) Create(context.Context, *user.User) error {
	return nil
}

func (r *fakeUserRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeUserRepository) FindByEmail(context.Context, string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepository) FindByEmailIncludingDeleted(context.Context, string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepository) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	account, ok := r.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return account, nil
}

func (r *fakeUserRepository) FindRoleByID(context.Context, uuid.UUID) (*user.Role, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepository) FindRoleByName(context.Context, string) (*user.Role, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepository) List(context.Context, user.ListUsersQuery) ([]user.User, int64, error) {
	return nil, 0, nil
}

func (r *fakeUserRepository) Update(context.Context, *user.User) error {
	return nil
}

type fakeProjectRepository struct {
	projects map[uuid.UUID]*project.Project
}

func (r *fakeProjectRepository) Create(context.Context, *project.Project) error {
	return nil
}

func (r *fakeProjectRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeProjectRepository) FindByCodeIncludingDeleted(context.Context, string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeProjectRepository) FindByID(_ context.Context, id uuid.UUID) (*project.Project, error) {
	project, ok := r.projects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return project, nil
}

func (r *fakeProjectRepository) List(context.Context, project.ListProjectsQuery) ([]project.Project, int64, error) {
	return nil, 0, nil
}

func (r *fakeProjectRepository) Update(context.Context, *project.Project) error {
	return nil
}

type fakeSprintRepository struct {
	sprints map[uuid.UUID]*sprint.Sprint
}

func (r *fakeSprintRepository) Create(context.Context, *sprint.Sprint) error {
	return nil
}

func (r *fakeSprintRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeSprintRepository) FindByID(_ context.Context, id uuid.UUID) (*sprint.Sprint, error) {
	sprint, ok := r.sprints[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return sprint, nil
}

func (r *fakeSprintRepository) List(context.Context, sprint.ListSprintsQuery) ([]sprint.Sprint, int64, error) {
	return nil, 0, nil
}

func (r *fakeSprintRepository) Update(context.Context, *sprint.Sprint) error {
	return nil
}

type fakeStatusRepository struct {
	statuses map[uuid.UUID]*status.TaskStatus
}

func (r *fakeStatusRepository) Create(context.Context, *status.TaskStatus) error {
	return nil
}

func (r *fakeStatusRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeStatusRepository) FindByID(_ context.Context, id uuid.UUID) (*status.TaskStatus, error) {
	taskStatus, ok := r.statuses[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return taskStatus, nil
}

func (r *fakeStatusRepository) FindByName(_ context.Context, name string) (*status.TaskStatus, error) {
	for _, taskStatus := range r.statuses {
		if strings.EqualFold(taskStatus.StatusName, name) {
			return taskStatus, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *fakeStatusRepository) List(context.Context, status.ListTaskStatusesQuery) ([]status.TaskStatus, int64, error) {
	return nil, 0, nil
}

func (r *fakeStatusRepository) Update(context.Context, *status.TaskStatus) error {
	return nil
}
