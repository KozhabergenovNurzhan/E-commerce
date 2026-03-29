package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/jmoiron/sqlx"
)

type AddressService struct {
	repo repository.AddressRepository
	db   *sqlx.DB
}

func NewAddressService(repo repository.AddressRepository, db *sqlx.DB) *AddressService {
	return &AddressService{repo: repo, db: db}
}

func (s *AddressService) Create(ctx context.Context, userID int64, req *models.CreateAddress) (*models.Address, error) {
	now := utils.Now()
	address := &models.Address{
		UserID:     userID,
		FullName:   req.FullName,
		Phone:      req.Phone,
		Country:    req.Country,
		City:       req.City,
		Street:     req.Street,
		PostalCode: req.PostalCode,
		IsDefault:  req.IsDefault,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if req.IsDefault {
		if err := s.repo.ClearDefault(ctx, userID); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Create(ctx, address); err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) List(ctx context.Context, userID int64) ([]*models.Address, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *AddressService) Update(ctx context.Context, addressID, userID int64, req *models.UpdateAddress) (*models.Address, error) {
	address, err := s.repo.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.UserID != userID {
		return nil, apperrors.Forbidden("cannot update another user's address", nil)
	}

	address.FullName = req.FullName
	address.Phone = req.Phone
	address.Country = req.Country
	address.City = req.City
	address.Street = req.Street
	address.PostalCode = req.PostalCode
	address.UpdatedAt = utils.Now()

	if err := s.repo.Update(ctx, address); err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) Delete(ctx context.Context, addressID, userID int64) error {
	address, err := s.repo.FindByID(ctx, addressID)
	if err != nil {
		return err
	}

	if address.UserID != userID {
		return apperrors.Forbidden("cannot delete another user's address", nil)
	}

	return s.repo.Delete(ctx, addressID)
}

func (s *AddressService) SetDefault(ctx context.Context, addressID, userID int64) (*models.Address, error) {
	address, err := s.repo.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.UserID != userID {
		return nil, apperrors.Forbidden("cannot update another user's address", nil)
	}

	if err := s.repo.ClearDefault(ctx, userID); err != nil {
		return nil, err
	}

	address.IsDefault = true
	address.UpdatedAt = utils.Now()

	const q = `UPDATE addresses SET is_default = true, updated_at = $1 WHERE id = $2`
	if _, err := s.db.ExecContext(ctx, q, address.UpdatedAt, addressID); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	return address, nil
}
