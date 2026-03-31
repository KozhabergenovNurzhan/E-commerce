package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.UserRecord) error
	FindByID(ctx context.Context, id int64) (*models.UserRecord, error)
	FindByEmail(ctx context.Context, email string) (*models.UserRecord, error)
	Update(ctx context.Context, user *models.UserRecord) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*models.UserRecord, int, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.UserRecord) error {
	const q = `
		INSERT INTO users (email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES (:email, :password_hash, :first_name, :last_name, :role, :created_at, :updated_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, user)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return apperrors.Conflict("email already taken", nil)
		}
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&user.ID)
	}

	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*models.UserRecord, error) {
	var u models.UserRecord

	const q = `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`

	if err := r.db.GetContext(ctx, &u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("user not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.UserRecord, error) {
	var u models.UserRecord

	const q = `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`

	if err := r.db.GetContext(ctx, &u, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("user not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.UserRecord) error {
	const q = `
		UPDATE users
		SET first_name = :first_name, last_name = :last_name, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, q, user)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("user not found", nil)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	const q = `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("user not found", nil)
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.UserRecord, int, error) {
	var users []*models.UserRecord

	const q = `SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	if err := r.db.SelectContext(ctx, &users, q, limit, offset); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	return users, total, nil
}
