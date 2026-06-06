package user

import (
	"context"
	"errors"
	"strings"

	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, filter ListUsersQuery) ([]UserResponse, map[string]any, error) {
	filter.Page = normalizePage(filter.Page)
	filter.Limit = normalizeLimit(filter.Limit)
	filter.Search = strings.TrimSpace(filter.Search)

	if filter.RoleID != "" {
		if _, err := uuid.Parse(filter.RoleID); err != nil {
			return nil, nil, apperrors.BadRequest("role_id must be a valid UUID")
		}
	}

	users, total, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]any{
		"page":  filter.Page,
		"limit": filter.Limit,
		"total": total,
	}

	return NewResponses(users), meta, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	account, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}

		return nil, err
	}

	response := NewResponse(*account)
	return &response, nil
}

func (s *Service) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, apperrors.BadRequest("role_id must be a valid UUID")
	}

	role, err := s.repository.FindRoleByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("role does not exist")
		}

		return nil, err
	}

	email := normalizeEmail(req.Email)
	if err := s.ensureEmailAvailable(ctx, email, uuid.Nil); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	account := &User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Role:         *role,
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		PasswordHash: passwordHash,
		Team:         strings.TrimSpace(req.Team),
		Position:     strings.TrimSpace(req.Position),
		IsActive:     isActive,
	}

	if err := s.repository.Create(ctx, account); err != nil {
		return nil, err
	}

	response := NewResponse(*account)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	account, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}

		return nil, err
	}

	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return nil, apperrors.BadRequest("role_id must be a valid UUID")
		}

		role, err := s.repository.FindRoleByID(ctx, roleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.BadRequest("role does not exist")
			}

			return nil, err
		}

		account.RoleID = role.ID
		account.Role = *role
	}

	if req.Name != nil {
		account.Name = strings.TrimSpace(*req.Name)
	}

	if req.Email != nil {
		email := normalizeEmail(*req.Email)
		if err := s.ensureEmailAvailable(ctx, email, account.ID); err != nil {
			return nil, err
		}

		account.Email = email
	}

	if req.Password != nil {
		passwordHash, err := hashPassword(*req.Password)
		if err != nil {
			return nil, err
		}

		account.PasswordHash = passwordHash
	}

	if req.Team != nil {
		account.Team = strings.TrimSpace(*req.Team)
	}

	if req.Position != nil {
		account.Position = strings.TrimSpace(*req.Position)
	}

	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	if err := s.repository.Update(ctx, account); err != nil {
		return nil, err
	}

	response := NewResponse(*account)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) ensureEmailAvailable(ctx context.Context, email string, currentID uuid.UUID) error {
	existing, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	if currentID == uuid.Nil || existing.ID != currentID {
		return apperrors.Conflict("email is already registered")
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}

	return page
}

func normalizeLimit(limit int) int {
	if limit < 1 {
		return 20
	}

	if limit > 100 {
		return 100
	}

	return limit
}
