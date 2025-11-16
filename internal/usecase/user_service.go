package usecase

import (
	"context"

	"prservice/internal/domain"
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) SetIsActive(ctx context.Context, id domain.UserID, active bool) (*domain.User, error) {
	u, err := s.users.SetIsActive(ctx, id, active)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domain.ErrNotFound
	}
	return u, nil
}
