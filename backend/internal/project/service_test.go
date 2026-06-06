package project

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestProjectCRUD(t *testing.T) {
	repository := newFakeProjectRepository()
	service := NewService(repository)

	created, err := service.Create(context.Background(), CreateProjectRequest{
		ProjectCode: " DEV ",
		ProjectName: " Dev Tracker ",
		ClientName:  " Internal ",
		Status:      "active",
		StartDate:   "2026-01-01",
		EndDate:     "2026-01-31",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if created.ProjectCode != "DEV" {
		t.Fatalf("expected trimmed project code, got %q", created.ProjectCode)
	}

	newName := "Developer Tracker"
	updated, err := service.Update(context.Background(), created.ID, UpdateProjectRequest{
		ProjectName: &newName,
	})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}
	if updated.ProjectName != newName {
		t.Fatalf("expected updated name %q, got %q", newName, updated.ProjectName)
	}

	list, meta, err := service.List(context.Background(), ListProjectsQuery{Page: 1, Limit: 10, Search: "dev"})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(list) != 1 || meta["total"].(int64) != 1 {
		t.Fatalf("expected one listed project, got len=%d meta=%v", len(list), meta)
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if !repository.deleted[created.ID] {
		t.Fatal("expected repository delete to be called")
	}
}

type fakeProjectRepository struct {
	projects map[uuid.UUID]*Project
	deleted  map[uuid.UUID]bool
}

func newFakeProjectRepository() *fakeProjectRepository {
	return &fakeProjectRepository{
		projects: map[uuid.UUID]*Project{},
		deleted:  map[uuid.UUID]bool{},
	}
}

func (r *fakeProjectRepository) Create(_ context.Context, project *Project) error {
	r.projects[project.ID] = project
	return nil
}

func (r *fakeProjectRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted[id] = true
	return nil
}

func (r *fakeProjectRepository) FindByCodeIncludingDeleted(_ context.Context, code string) (*Project, error) {
	for _, project := range r.projects {
		if project.ProjectCode == code {
			return project, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeProjectRepository) FindByID(_ context.Context, id uuid.UUID) (*Project, error) {
	project, ok := r.projects[id]
	if !ok || r.deleted[id] {
		return nil, gorm.ErrRecordNotFound
	}
	return project, nil
}

func (r *fakeProjectRepository) List(context.Context, ListProjectsQuery) ([]Project, int64, error) {
	projects := make([]Project, 0, len(r.projects))
	for id, project := range r.projects {
		if !r.deleted[id] {
			projects = append(projects, *project)
		}
	}
	return projects, int64(len(projects)), nil
}

func (r *fakeProjectRepository) Update(_ context.Context, project *Project) error {
	r.projects[project.ID] = project
	return nil
}
