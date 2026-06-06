package status

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestTaskStatusCRUD(t *testing.T) {
	repository := newFakeStatusRepository()
	service := NewService(repository)

	created, err := service.Create(context.Background(), CreateTaskStatusRequest{
		StatusName:  " Review ",
		ColorName:   " purple ",
		ColorHex:    "#a855f7",
		StatusOrder: 7,
		IsActive:    boolPtr(true),
	})
	if err != nil {
		t.Fatalf("create status: %v", err)
	}
	if created.ColorHex != "#A855F7" {
		t.Fatalf("expected normalized color hex, got %q", created.ColorHex)
	}

	isDone := true
	updated, err := service.Update(context.Background(), created.ID, UpdateTaskStatusRequest{
		IsDone: &isDone,
	})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if !updated.IsDone {
		t.Fatal("expected status to be done")
	}

	if _, _, err := service.List(context.Background(), ListTaskStatusesQuery{Page: 1, Limit: 20}); err != nil {
		t.Fatalf("list statuses: %v", err)
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete status: %v", err)
	}
	if !repository.deleted[created.ID] {
		t.Fatal("expected repository delete to be called")
	}
}

func boolPtr(value bool) *bool {
	return &value
}

type fakeStatusRepository struct {
	statuses map[uuid.UUID]*TaskStatus
	deleted  map[uuid.UUID]bool
}

func newFakeStatusRepository() *fakeStatusRepository {
	return &fakeStatusRepository{
		statuses: map[uuid.UUID]*TaskStatus{},
		deleted:  map[uuid.UUID]bool{},
	}
}

func (r *fakeStatusRepository) Create(_ context.Context, taskStatus *TaskStatus) error {
	r.statuses[taskStatus.ID] = taskStatus
	return nil
}

func (r *fakeStatusRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted[id] = true
	return nil
}

func (r *fakeStatusRepository) FindByID(_ context.Context, id uuid.UUID) (*TaskStatus, error) {
	taskStatus, ok := r.statuses[id]
	if !ok || r.deleted[id] {
		return nil, gorm.ErrRecordNotFound
	}
	return taskStatus, nil
}

func (r *fakeStatusRepository) FindByName(_ context.Context, name string) (*TaskStatus, error) {
	for _, taskStatus := range r.statuses {
		if taskStatus.StatusName == name {
			return taskStatus, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeStatusRepository) List(context.Context, ListTaskStatusesQuery) ([]TaskStatus, int64, error) {
	statuses := make([]TaskStatus, 0, len(r.statuses))
	for id, taskStatus := range r.statuses {
		if !r.deleted[id] {
			statuses = append(statuses, *taskStatus)
		}
	}
	return statuses, int64(len(statuses)), nil
}

func (r *fakeStatusRepository) Update(_ context.Context, taskStatus *TaskStatus) error {
	r.statuses[taskStatus.ID] = taskStatus
	return nil
}
