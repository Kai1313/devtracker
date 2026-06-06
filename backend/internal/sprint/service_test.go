package sprint

import (
	"context"
	"testing"

	"devtracker/backend/internal/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestSprintCRUD(t *testing.T) {
	projectID := uuid.New()
	projectRepository := &fakeProjectRepository{
		projects: map[uuid.UUID]*project.Project{
			projectID: {ID: projectID, ProjectCode: "DEV", ProjectName: "Dev Tracker"},
		},
	}
	sprintRepository := newFakeSprintRepository()
	service := NewService(sprintRepository, projectRepository)

	created, err := service.Create(context.Background(), CreateSprintRequest{
		ProjectID:  projectID.String(),
		SprintName: " Sprint 1 ",
		StartDate:  "2026-01-01",
		EndDate:    "2026-01-14",
		Status:     StatusActive,
	})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	if created.Status != StatusActive {
		t.Fatalf("expected active status, got %q", created.Status)
	}

	closed, err := service.Close(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("close sprint: %v", err)
	}
	if closed.Status != StatusClosed {
		t.Fatalf("expected closed status, got %q", closed.Status)
	}

	newName := "Sprint 1A"
	updated, err := service.Update(context.Background(), created.ID, UpdateSprintRequest{
		SprintName: &newName,
	})
	if err != nil {
		t.Fatalf("update sprint: %v", err)
	}
	if updated.SprintName != newName {
		t.Fatalf("expected updated sprint name %q, got %q", newName, updated.SprintName)
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete sprint: %v", err)
	}
	if !sprintRepository.deleted[created.ID] {
		t.Fatal("expected repository delete to be called")
	}
}

type fakeSprintRepository struct {
	sprints map[uuid.UUID]*Sprint
	deleted map[uuid.UUID]bool
}

func newFakeSprintRepository() *fakeSprintRepository {
	return &fakeSprintRepository{
		sprints: map[uuid.UUID]*Sprint{},
		deleted: map[uuid.UUID]bool{},
	}
}

func (r *fakeSprintRepository) Create(_ context.Context, sprint *Sprint) error {
	r.sprints[sprint.ID] = sprint
	return nil
}

func (r *fakeSprintRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted[id] = true
	return nil
}

func (r *fakeSprintRepository) FindByID(_ context.Context, id uuid.UUID) (*Sprint, error) {
	sprint, ok := r.sprints[id]
	if !ok || r.deleted[id] {
		return nil, gorm.ErrRecordNotFound
	}
	return sprint, nil
}

func (r *fakeSprintRepository) List(context.Context, ListSprintsQuery) ([]Sprint, int64, error) {
	sprints := make([]Sprint, 0, len(r.sprints))
	for id, sprint := range r.sprints {
		if !r.deleted[id] {
			sprints = append(sprints, *sprint)
		}
	}
	return sprints, int64(len(sprints)), nil
}

func (r *fakeSprintRepository) Update(_ context.Context, sprint *Sprint) error {
	r.sprints[sprint.ID] = sprint
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
