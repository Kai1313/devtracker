package user

import (
	"context"
	"strings"

	appquery "devtracker/backend/internal/query"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Count(ctx context.Context) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	Create(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByEmailIncludingDeleted(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindRoleByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context, filter ListUsersQuery) ([]User, int64, error)
	Update(ctx context.Context, user *User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Count(&count).Error
	return count, err
}

func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&User{}).Count(&count).Error
	return count, err
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Role", "Roles", "Role.Permissions").Create(user).Error; err != nil {
			return err
		}

		return syncPrimaryUserRole(tx, user.ID, user.RoleID)
	})
}

func (r *repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Role", "Roles", "Role.Permissions").Save(user).Error; err != nil {
			return err
		}

		return syncPrimaryUserRole(tx, user.ID, user.RoleID)
	})
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&User{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Preload("Roles.Permissions").
		First(&user, "id = ?", id).
		Error

	return &user, err
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Preload("Roles.Permissions").
		First(&user, "LOWER(email) = ?", strings.ToLower(email)).
		Error

	return &user, err
}

func (r *repository) FindByEmailIncludingDeleted(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).
		Unscoped().
		Preload("Role.Permissions").
		Preload("Roles.Permissions").
		First(&user, "LOWER(email) = ?", strings.ToLower(email)).
		Error

	return &user, err
}

func (r *repository) FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).
		Error

	return &role, err
}

func (r *repository) FindRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "LOWER(name) = ?", strings.ToLower(name)).
		Error

	return &role, err
}

func (r *repository) List(ctx context.Context, filter ListUsersQuery) ([]User, int64, error) {
	query := r.db.WithContext(ctx).Model(&User{}).Preload("Role")

	if filter.Search != "" {
		term := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", term, term)
	}

	if filter.RoleID != "" {
		query = query.Where("role_id = ?", filter.RoleID)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []User
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Order(appquery.OrderClause(appquery.Sort{By: filter.SortBy, Order: filter.SortOrder}, userSortFields)).
		Offset(offset).
		Limit(filter.Limit).
		Find(&users).
		Error

	return users, total, err
}

func syncPrimaryUserRole(tx *gorm.DB, userID, roleID uuid.UUID) error {
	if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
		return err
	}

	return tx.Exec(
		"INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		userID,
		roleID,
	).Error
}
