package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestRecordCreatesAuditLog(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	userID := uuid.New()

	err := service.Record(context.Background(), RecordInput{
		UserID:    &userID,
		Module:    " Projects ",
		Action:    " Create ",
		NewValue:  map[string]any{"project_code": "DEV"},
		IPAddress: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("record audit log: %v", err)
	}

	if repository.created == nil {
		t.Fatal("expected audit log to be created")
	}
	if repository.created.UserID == nil || *repository.created.UserID != userID {
		t.Fatalf("expected user_id %s, got %v", userID, repository.created.UserID)
	}
	if repository.created.Module != "projects" {
		t.Fatalf("expected normalized module projects, got %q", repository.created.Module)
	}
	if repository.created.Action != "create" {
		t.Fatalf("expected normalized action create, got %q", repository.created.Action)
	}
	if repository.created.NewValue == nil || *repository.created.NewValue != `{"project_code":"DEV"}` {
		t.Fatalf("expected JSON new_value, got %v", repository.created.NewValue)
	}
}

func TestListNormalizesFilters(t *testing.T) {
	userID := uuid.New()
	repository := &fakeRepository{}
	service := NewService(repository)

	_, meta, err := service.List(context.Background(), ListQuery{
		Page:      0,
		Limit:     200,
		UserID:    userID.String(),
		Module:    " Tasks ",
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
	})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}

	if meta["page"].(int) != 1 {
		t.Fatalf("expected normalized page 1, got %v", meta["page"])
	}
	if meta["limit"].(int) != 100 {
		t.Fatalf("expected normalized limit 100, got %v", meta["limit"])
	}
	if repository.filter.UserID == nil || *repository.filter.UserID != userID {
		t.Fatalf("expected user filter %s, got %v", userID, repository.filter.UserID)
	}
	if repository.filter.Module != "tasks" {
		t.Fatalf("expected normalized module tasks, got %q", repository.filter.Module)
	}
	if repository.filter.StartDate == nil || repository.filter.StartDate.Format(dateLayout) != "2026-01-01" {
		t.Fatalf("expected start_date 2026-01-01, got %v", repository.filter.StartDate)
	}
	if repository.filter.EndDate == nil || repository.filter.EndDate.Format(dateLayout) != "2026-02-01" {
		t.Fatalf("expected exclusive end_date 2026-02-01, got %v", repository.filter.EndDate)
	}
}

type fakeRepository struct {
	created *AuditLog
	filter  listFilter
}

func (r *fakeRepository) Create(_ context.Context, log *AuditLog) error {
	r.created = log
	return nil
}

func (r *fakeRepository) List(_ context.Context, filter listFilter) ([]AuditLog, int64, error) {
	r.filter = filter
	return nil, 0, nil
}
