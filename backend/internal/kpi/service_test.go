package kpi

import (
	"context"
	"testing"

	"devtracker/backend/internal/sprint"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestDevelopersUsesSnapshotsForClosedSprint(t *testing.T) {
	sprintID := uuid.New()
	repository := &fakeKPIRepository{}
	service := NewService(repository, &fakeSprintRepository{
		sprints: map[uuid.UUID]*sprint.Sprint{
			sprintID: {ID: sprintID, Status: sprint.StatusClosed},
		},
	})

	result, err := service.Developers(context.Background(), Query{SprintID: sprintID.String()})
	if err != nil {
		t.Fatalf("developer KPI: %v", err)
	}

	if !repository.developerSnapshotCalled {
		t.Fatal("expected snapshot developer KPI to be used")
	}
	if repository.developerLiveCalled {
		t.Fatal("did not expect live developer KPI to be used")
	}
	if len(result) != 1 || result[0].DeveloperName != "snapshot" {
		t.Fatalf("expected snapshot result, got %+v", result)
	}
}

func TestProjectsUsesLiveKPIForOpenSprint(t *testing.T) {
	sprintID := uuid.New()
	repository := &fakeKPIRepository{}
	service := NewService(repository, &fakeSprintRepository{
		sprints: map[uuid.UUID]*sprint.Sprint{
			sprintID: {ID: sprintID, Status: sprint.StatusActive},
		},
	})

	result, err := service.Projects(context.Background(), Query{SprintID: sprintID.String()})
	if err != nil {
		t.Fatalf("project KPI: %v", err)
	}

	if !repository.projectLiveCalled {
		t.Fatal("expected live project KPI to be used")
	}
	if repository.projectSnapshotCalled {
		t.Fatal("did not expect snapshot project KPI to be used")
	}
	if len(result) != 1 || result[0].ProjectName != "live" {
		t.Fatalf("expected live result, got %+v", result)
	}
}

func TestGenerateSprintSnapshotsDelegatesToRepository(t *testing.T) {
	sprintID := uuid.New()
	repository := &fakeKPIRepository{}
	service := NewService(repository, &fakeSprintRepository{})

	if err := service.GenerateSprintSnapshots(context.Background(), sprintID); err != nil {
		t.Fatalf("generate sprint snapshots: %v", err)
	}

	if repository.generatedSprintID == nil || *repository.generatedSprintID != sprintID {
		t.Fatalf("expected generated sprint %s, got %v", sprintID, repository.generatedSprintID)
	}
}

type fakeKPIRepository struct {
	developerLiveCalled     bool
	developerSnapshotCalled bool
	projectLiveCalled       bool
	projectSnapshotCalled   bool
	generatedSprintID       *uuid.UUID
}

func (r *fakeKPIRepository) DeveloperKPI(context.Context, *uuid.UUID) ([]DeveloperKPIResponse, error) {
	r.developerLiveCalled = true
	return []DeveloperKPIResponse{{DeveloperName: "live"}}, nil
}

func (r *fakeKPIRepository) DeveloperSnapshotKPI(context.Context, uuid.UUID) ([]DeveloperKPIResponse, error) {
	r.developerSnapshotCalled = true
	return []DeveloperKPIResponse{{DeveloperName: "snapshot"}}, nil
}

func (r *fakeKPIRepository) GenerateSprintSnapshots(_ context.Context, sprintID uuid.UUID) error {
	r.generatedSprintID = &sprintID
	return nil
}

func (r *fakeKPIRepository) ProjectKPI(context.Context, *uuid.UUID) ([]ProjectKPIResponse, error) {
	r.projectLiveCalled = true
	return []ProjectKPIResponse{{ProjectName: "live"}}, nil
}

func (r *fakeKPIRepository) ProjectSnapshotKPI(context.Context, uuid.UUID) ([]ProjectKPIResponse, error) {
	r.projectSnapshotCalled = true
	return []ProjectKPIResponse{{ProjectName: "snapshot"}}, nil
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
