package workload

import (
	"context"
	"testing"

	"devtracker/backend/internal/project"
	"devtracker/backend/internal/sprint"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestDeveloperWorkloadAppliesFilters(t *testing.T) {
	sprintID := uuid.New()
	projectID := uuid.New()
	repository := &fakeWorkloadRepository{}
	service := NewService(
		repository,
		&fakeSprintRepository{sprints: map[uuid.UUID]*sprint.Sprint{
			sprintID: {ID: sprintID, ProjectID: projectID, Status: sprint.StatusActive},
		}},
		&fakeProjectRepository{projects: map[uuid.UUID]*project.Project{
			projectID: {ID: projectID, ProjectCode: "DEV", ProjectName: "Dev Tracker"},
		}},
	)

	result, err := service.DeveloperWorkload(context.Background(), Query{
		SprintID:  sprintID.String(),
		ProjectID: projectID.String(),
	})
	if err != nil {
		t.Fatalf("developer workload: %v", err)
	}

	if repository.filter.SprintID == nil || *repository.filter.SprintID != sprintID {
		t.Fatalf("expected sprint filter %s, got %v", sprintID, repository.filter.SprintID)
	}
	if repository.filter.ProjectID == nil || *repository.filter.ProjectID != projectID {
		t.Fatalf("expected project filter %s, got %v", projectID, repository.filter.ProjectID)
	}
	if len(result) != 1 || result[0].WorkloadClassification != ClassificationHigh {
		t.Fatalf("expected high workload result, got %+v", result)
	}
}

func TestDeveloperWorkloadRejectsInvalidSprint(t *testing.T) {
	service := NewService(
		&fakeWorkloadRepository{},
		&fakeSprintRepository{},
		&fakeProjectRepository{},
	)

	if _, err := service.DeveloperWorkload(context.Background(), Query{SprintID: "bad-id"}); err == nil {
		t.Fatal("expected invalid sprint to fail")
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name   string
		points float64
		want   string
	}{
		{name: "low", points: 4.99, want: ClassificationLow},
		{name: "normal", points: 5, want: ClassificationNormal},
		{name: "high", points: 13.5, want: ClassificationHigh},
		{name: "overloaded", points: 20.5, want: ClassificationOverloaded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classify(tt.points); got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

type fakeWorkloadRepository struct {
	filter filter
}

func (r *fakeWorkloadRepository) DeveloperWorkload(_ context.Context, filter filter) ([]DeveloperWorkloadResponse, error) {
	r.filter = filter
	return []DeveloperWorkloadResponse{
		{
			DeveloperID:            uuid.New(),
			DeveloperName:          "Dev User",
			ActiveTasks:            3,
			TotalPoints:            14,
			OverdueTasks:           1,
			CurrentSprintTasks:     2,
			WorkloadClassification: classify(14),
		},
	}, nil
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
	current, ok := r.sprints[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}

	return current, nil
}

func (r *fakeSprintRepository) List(context.Context, sprint.ListSprintsQuery) ([]sprint.Sprint, int64, error) {
	return nil, 0, nil
}

func (r *fakeSprintRepository) Update(context.Context, *sprint.Sprint) error {
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
	current, ok := r.projects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}

	return current, nil
}

func (r *fakeProjectRepository) List(context.Context, project.ListProjectsQuery) ([]project.Project, int64, error) {
	return nil, 0, nil
}

func (r *fakeProjectRepository) Update(context.Context, *project.Project) error {
	return nil
}
