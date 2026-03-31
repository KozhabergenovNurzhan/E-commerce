package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"golang.org/x/crypto/bcrypt"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, req *models.Register) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	now := utils.Now()
	user := &models.UserRecord{
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RoleCustomer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

// Login validates credentials and returns the full UserRecord so the caller can
// generate tokens with the user's ID and role.
func (s *UserService) Login(ctx context.Context, req *models.Login) (*models.UserRecord, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.BadRequest("invalid credentials", nil) // avoid email enumeration
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.BadRequest("invalid credentials", nil)
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *UserService) Update(ctx context.Context, id int64, firstName, lastName string) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = utils.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) List(ctx context.Context, page, limit int) ([]*models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	users, total, err := s.repo.List(ctx, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*models.User, len(users))
	for i, u := range users {
		resp[i] = u.ToResponse()
	}
	return resp, total, nil
}
