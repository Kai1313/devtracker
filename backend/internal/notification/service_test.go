package notification

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreateNotification(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository)
	userID := uuid.New()
	taskID := uuid.New()

	err := service.Create(context.Background(), CreateInput{
		UserID:  userID,
		TaskID:  &taskID,
		Type:    " Task_Assigned ",
		Title:   " Task assigned ",
		Message: " Build API ",
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if len(repository.notifications) != 1 {
		t.Fatalf("expected one notification, got %d", len(repository.notifications))
	}

	created := repository.notifications[0]
	if created.UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, created.UserID)
	}
	if created.TaskID == nil || *created.TaskID != taskID {
		t.Fatalf("expected task_id %s, got %v", taskID, created.TaskID)
	}
	if created.Type != TypeTaskAssigned {
		t.Fatalf("expected type %q, got %q", TypeTaskAssigned, created.Type)
	}
	if created.Title != "Task assigned" {
		t.Fatalf("expected trimmed title, got %q", created.Title)
	}
	if created.Message != "Build API" {
		t.Fatalf("expected trimmed message, got %q", created.Message)
	}
	if created.IsRead {
		t.Fatal("expected notification to start unread")
	}
}

func TestListReturnsUnreadCount(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository)
	userID := uuid.New()
	otherUserID := uuid.New()

	repository.notifications = []Notification{
		{ID: uuid.New(), UserID: userID, Type: TypeTaskAssigned, Title: "A", Message: "A"},
		{ID: uuid.New(), UserID: userID, Type: TypeTaskDone, Title: "B", Message: "B", IsRead: true},
		{ID: uuid.New(), UserID: otherUserID, Type: TypeTaskAssigned, Title: "C", Message: "C"},
	}

	result, meta, err := service.List(context.Background(), ListQuery{
		Page:   0,
		Limit:  200,
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}

	if result.UnreadCount != 1 {
		t.Fatalf("expected unread count 1, got %d", result.UnreadCount)
	}
	if len(result.Notifications) != 2 {
		t.Fatalf("expected two user notifications, got %d", len(result.Notifications))
	}
	if meta["page"].(int) != 1 {
		t.Fatalf("expected normalized page 1, got %v", meta["page"])
	}
	if meta["limit"].(int) != 100 {
		t.Fatalf("expected normalized limit 100, got %v", meta["limit"])
	}
	if meta["total"].(int64) != 2 {
		t.Fatalf("expected total 2, got %v", meta["total"])
	}
}

func TestMarkReadUpdatesUnreadCount(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository)
	userID := uuid.New()
	notificationID := uuid.New()

	repository.notifications = []Notification{
		{ID: notificationID, UserID: userID, Type: TypeTaskDone, Title: "Done", Message: "Done"},
	}

	result, err := service.MarkRead(context.Background(), notificationID, userID)
	if err != nil {
		t.Fatalf("mark read: %v", err)
	}

	if !result.Notification.IsRead {
		t.Fatal("expected notification to be marked read")
	}
	if result.Notification.ReadAt == nil {
		t.Fatal("expected read_at to be set")
	}
	if result.UnreadCount != 0 {
		t.Fatalf("expected unread count 0, got %d", result.UnreadCount)
	}
}

type fakeRepository struct {
	notifications []Notification
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{notifications: []Notification{}}
}

func (r *fakeRepository) CountUnread(_ context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	for _, notification := range r.notifications {
		if notification.UserID == userID && !notification.IsRead {
			count++
		}
	}

	return count, nil
}

func (r *fakeRepository) Create(_ context.Context, notification *Notification) error {
	r.notifications = append(r.notifications, *notification)
	return nil
}

func (r *fakeRepository) FindByIDForUser(_ context.Context, id uuid.UUID, userID uuid.UUID) (*Notification, error) {
	for i := range r.notifications {
		if r.notifications[i].ID == id && r.notifications[i].UserID == userID {
			return &r.notifications[i], nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *fakeRepository) List(_ context.Context, query ListQuery) ([]Notification, int64, error) {
	result := make([]Notification, 0, len(r.notifications))
	for _, notification := range r.notifications {
		if notification.UserID == query.UserID {
			result = append(result, notification)
		}
	}

	return result, int64(len(result)), nil
}

func (r *fakeRepository) Update(_ context.Context, notification *Notification) error {
	for i := range r.notifications {
		if r.notifications[i].ID == notification.ID {
			r.notifications[i] = *notification
			return nil
		}
	}

	return gorm.ErrRecordNotFound
}
