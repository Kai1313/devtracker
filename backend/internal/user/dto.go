package user

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	RoleID   string `json:"role_id" validate:"required,uuid"`
	Name     string `json:"name" validate:"required,min=2,max=150"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Team     string `json:"team" validate:"omitempty,max=100"`
	Position string `json:"position" validate:"omitempty,max=100"`
	IsActive *bool  `json:"is_active" validate:"omitempty"`
}

type UpdateUserRequest struct {
	RoleID   *string `json:"role_id" validate:"omitempty,uuid"`
	Name     *string `json:"name" validate:"omitempty,min=2,max=150"`
	Email    *string `json:"email" validate:"omitempty,email,max=150"`
	Password *string `json:"password" validate:"omitempty,min=8,max=72"`
	Team     *string `json:"team" validate:"omitempty,max=100"`
	Position *string `json:"position" validate:"omitempty,max=100"`
	IsActive *bool   `json:"is_active" validate:"omitempty"`
}

type ListUsersQuery struct {
	Page     int
	Limit    int
	Search   string
	RoleID   string
	IsActive *bool
}

type UserResponse struct {
	ID        uuid.UUID    `json:"id"`
	RoleID    uuid.UUID    `json:"role_id"`
	Role      RoleResponse `json:"role"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Team      string       `json:"team,omitempty"`
	Position  string       `json:"position,omitempty"`
	IsActive  bool         `json:"is_active"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

func NewResponse(model User) UserResponse {
	return UserResponse{
		ID:        model.ID,
		RoleID:    model.RoleID,
		Role:      NewRoleResponse(model.Role),
		Name:      model.Name,
		Email:     model.Email,
		Team:      model.Team,
		Position:  model.Position,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func NewResponses(models []User) []UserResponse {
	result := make([]UserResponse, 0, len(models))
	for _, model := range models {
		result = append(result, NewResponse(model))
	}

	return result
}

func NewRoleResponse(model Role) RoleResponse {
	return RoleResponse{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
	}
}
