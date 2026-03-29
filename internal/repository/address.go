package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"
)

type AddressRepository interface {
	Create(ctx context.Context, a *models.Address) error
	FindByID(ctx context.Context, id int64) (*models.Address, error)
	ListByUser(ctx context.Context, userID int64) ([]*models.Address, error)
	Update(ctx context.Context, a *models.Address) error
	Delete(ctx context.Context, id int64) error
	ClearDefault(ctx context.Context, userID int64) error
}

type addressRepository struct {
	db *sqlx.DB
}

func NewAddressRepository(db *sqlx.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, a *models.Address) error {
	const q = `
		INSERT INTO addresses (user_id, full_name, phone, country, city, street, postal_code, is_default, created_at, updated_at)
		VALUES (:user_id, :full_name, :phone, :country, :city, :street, :postal_code, :is_default, :created_at, :updated_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, a)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&a.ID)
	}

	return nil
}

func (r *addressRepository) FindByID(ctx context.Context, id int64) (*models.Address, error) {
	var a models.Address

	if err := r.db.GetContext(ctx, &a, `SELECT * FROM addresses WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("address not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	return &a, nil
}

func (r *addressRepository) ListByUser(ctx context.Context, userID int64) ([]*models.Address, error) {
	var addresses []*models.Address

	const q = `SELECT * FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`

	if err := r.db.SelectContext(ctx, &addresses, q, userID); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	return addresses, nil
}

func (r *addressRepository) Update(ctx context.Context, a *models.Address) error {
	const q = `
		UPDATE addresses
		SET full_name = :full_name, phone = :phone, country = :country,
		    city = :city, street = :street, postal_code = :postal_code, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, q, a)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("address not found", nil)
	}

	return nil
}

func (r *addressRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, id)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("address not found", nil)
	}

	return nil
}

func (r *addressRepository) ClearDefault(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1`, userID)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	return nil
}
