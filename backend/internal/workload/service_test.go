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
	developerID := uuid.New()
	statusID := uuid.New()
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
		SprintID:    sprintID.String(),
		ProjectID:   projectID.String(),
		DeveloperID: developerID.String(),
		StatusID:    statusID.String(),
		StartDate:   "2026-01-01",
		EndDate:     "2026-01-31",
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
	if repository.filter.DeveloperID == nil || *repository.filter.DeveloperID != developerID {
		t.Fatalf("expected developer filter %s, got %v", developerID, repository.filter.DeveloperID)
	}
	if repository.filter.StatusID == nil || *repository.filter.StatusID != statusID {
		t.Fatalf("expected status filter %s, got %v", statusID, repository.filter.StatusID)
	}
	if repository.filter.StartDate == nil || repository.filter.StartDate.Format(dateLayout) != "2026-01-01" {
		t.Fatalf("expected start date filter, got %v", repository.filter.StartDate)
	}
	if repository.filter.EndDate == nil || repository.filter.EndDate.Format(dateLayout) != "2026-01-31" {
		t.Fatalf("expected end date filter, got %v", repository.filter.EndDate)
	}
	if len(result) != 1 || result[0].WorkloadLevel != ClassificationHigh {
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

func TestDeveloperWorkloadScopeRestrictsDeveloper(t *testing.T) {
	developerID := uuid.New()
	otherDeveloperID := uuid.New()
	repository := &fakeWorkloadRepository{}
	service := NewService(
		repository,
		&fakeSprintRepository{},
		&fakeProjectRepository{},
	)

	if _, err := service.DeveloperWorkloadWithScope(context.Background(), Query{
		DeveloperID: otherDeveloperID.String(),
	}, AccessScope{UserID: developerID, IsDeveloper: true}); err == nil {
		t.Fatal("expected developer to be forbidden from other developer workload")
	}

	if _, err := service.DeveloperWorkloadWithScope(context.Background(), Query{}, AccessScope{
		UserID:      developerID,
		IsDeveloper: true,
	}); err != nil {
		t.Fatalf("expected own workload scope to be allowed: %v", err)
	}

	if repository.filter.DeveloperID == nil || *repository.filter.DeveloperID != developerID {
		t.Fatalf("expected developer scope %s, got %v", developerID, repository.filter.DeveloperID)
	}
}

func TestDeveloperWorkloadScopeRestrictsQAStatuses(t *testing.T) {
	repository := &fakeWorkloadRepository{}
	service := NewService(
		repository,
		&fakeSprintRepository{},
		&fakeProjectRepository{},
	)

	if _, err := service.DeveloperWorkloadWithScope(context.Background(), Query{}, AccessScope{IsQA: true}); err != nil {
		t.Fatalf("expected QA scope to be allowed: %v", err)
	}

	if !repository.filter.QAOnly {
		t.Fatal("expected QA scope to apply QA-only filter")
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name        string
		activeTasks int64
		want        string
	}{
		{name: "low", activeTasks: 3, want: ClassificationLow},
		{name: "normal", activeTasks: 4, want: ClassificationNormal},
		{name: "high", activeTasks: 8, want: ClassificationHigh},
		{name: "overloaded", activeTasks: 11, want: ClassificationOverloaded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classify(tt.activeTasks); got != tt.want {
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
			DeveloperID:          uuid.New(),
			DeveloperName:        "Dev User",
			ActiveTasks:          8,
			DoneTasks:            3,
			OverdueTasks:         1,
			TotalEstimatedPoints: 14,
			TotalActualPoints:    10,
			CurrentSprintTasks:   2,
			WorkloadScore:        8,
			WorkloadLevel:        classify(8),
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
