package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type UserService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, limit int) ([]*domain.UserResponse, int, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         domain.RoleCustomer,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

// Login validates credentials and returns the full User so the caller can
// generate tokens with the user's ID and role.
func (s *userService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrBadRequest // avoid email enumeration
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrBadRequest
	}
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) List(ctx context.Context, page, limit int) ([]*domain.UserResponse, int, error) {
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

	resp := make([]*domain.UserResponse, len(users))
	for i, u := range users {
		resp[i] = u.ToResponse()
	}
	return resp, total, nil
}
