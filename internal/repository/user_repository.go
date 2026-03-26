package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.User, int, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :first_name, :last_name, :role, :is_active, :created_at, :updated_at)`

	if _, err := r.db.NamedExecContext(ctx, q, user); err != nil {
		if strings.Contains(err.Error(), "23505") {
			return apperrors.ErrConflict
		}
		return apperrors.ErrInternal
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	const q = `SELECT * FROM users WHERE id = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, &u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.ErrInternal
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	const q = `SELECT * FROM users WHERE email = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, &u, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.ErrInternal
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	const q = `
		UPDATE users
		SET first_name = :first_name, last_name = :last_name, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, q, user)
	if err != nil {
		return apperrors.ErrInternal
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1 AND is_active = true`
	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.ErrInternal
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	var users []*domain.User
	const q = `SELECT * FROM users WHERE is_active = true ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &users, q, limit, offset); err != nil {
		return nil, 0, apperrors.ErrInternal
	}
	var total int
	_ = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users WHERE is_active = true`)
	return users, total, nil
}
