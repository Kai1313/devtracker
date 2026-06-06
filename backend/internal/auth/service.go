package auth

import (
	"context"
	"errors"
	"strings"

	"devtracker/backend/internal/config"
	"devtracker/backend/internal/user"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	users  user.Repository
	tokens *TokenManager
	ttl    int64
}

func NewService(users user.Repository, jwtConfig config.JWTConfig) *Service {
	return &Service{
		users:  users,
		tokens: NewTokenManager(jwtConfig),
		ttl:    int64(jwtConfig.AccessTokenTTL.Seconds()),
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	account, err := s.users.FindByEmail(ctx, normalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Unauthorized("invalid email or password")
		}

		return nil, err
	}

	if !account.IsActive {
		return nil, apperrors.Forbidden("user account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	token, expiresAt, err := s.tokens.Generate(TokenInput{
		UserID: account.ID.String(),
		Email:  account.Email,
		Name:   account.Name,
		Role:   account.Role.Name,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt.Format(timeFormatRFC3339),
		ExpiresIn:   s.ttl,
		User:        user.NewResponse(*account),
	}, nil
}

func (s *Service) BootstrapAdmin(ctx context.Context, req BootstrapAdminRequest) (*user.UserResponse, error) {
	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, apperrors.Conflict("bootstrap admin can only be created before any users exist")
	}

	role, err := s.users.FindRoleByName(ctx, "admin")
	if err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	account := &user.User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Name:         strings.TrimSpace(req.Name),
		Email:        normalizeEmail(req.Email),
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	if err := s.users.Create(ctx, account); err != nil {
		return nil, err
	}

	account.Role = *role
	response := user.NewResponse(*account)

	return &response, nil
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*user.UserResponse, error) {
	account, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}

		return nil, err
	}

	response := user.NewResponse(*account)
	return &response, nil
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

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
