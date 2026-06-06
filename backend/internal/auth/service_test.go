package auth

import (
	"context"
	"testing"
	"time"

	"devtracker/backend/internal/config"
	"devtracker/backend/internal/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestLoginSuccess(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	role := user.Role{ID: uuid.New(), Name: "admin"}
	account := &user.User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Role:         role,
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: string(passwordHash),
		IsActive:     true,
	}

	service := NewService(&fakeUserRepository{usersByEmail: map[string]*user.User{
		account.Email: account,
	}}, config.JWTConfig{
		Secret:         "test-secret",
		Issuer:         "test",
		AccessTokenTTL: time.Hour,
	})

	result, err := service.Login(context.Background(), LoginRequest{
		Email:    " ADMIN@example.com ",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.User.ID != account.ID {
		t.Fatalf("expected user %s, got %s", account.ID, result.User.ID)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	role := user.Role{ID: uuid.New(), Name: "admin"}
	account := &user.User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Role:         role,
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: string(passwordHash),
		IsActive:     true,
	}

	service := NewService(&fakeUserRepository{usersByEmail: map[string]*user.User{
		account.Email: account,
	}}, config.JWTConfig{
		Secret:         "test-secret",
		Issuer:         "test",
		AccessTokenTTL: time.Hour,
	})

	if _, err := service.Login(context.Background(), LoginRequest{
		Email:    "admin@example.com",
		Password: "bad-password",
	}); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}

type fakeUserRepository struct {
	usersByEmail map[string]*user.User
}

func (r *fakeUserRepository) Count(context.Context) (int64, error) {
	return int64(len(r.usersByEmail)), nil
}

func (r *fakeUserRepository) CountAll(context.Context) (int64, error) {
	return int64(len(r.usersByEmail)), nil
}

func (r *fakeUserRepository) Create(_ context.Context, account *user.User) error {
	if r.usersByEmail == nil {
		r.usersByEmail = map[string]*user.User{}
	}
	r.usersByEmail[account.Email] = account
	return nil
}

func (r *fakeUserRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	account, ok := r.usersByEmail[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return account, nil
}

func (r *fakeUserRepository) FindByEmailIncludingDeleted(ctx context.Context, email string) (*user.User, error) {
	return r.FindByEmail(ctx, email)
}

func (r *fakeUserRepository) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	for _, account := range r.usersByEmail {
		if account.ID == id {
			return account, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
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
