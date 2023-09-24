package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
)

// UserService acts as a struct for injecting an implementation of UserRepository
// for use in service methods
type UserService struct {
	UserRepository domain.UserRepository
}

// USConfig will hold repositories that will eventually be injected into this
//
//	service layer
type USConfig struct {
	UserRepository domain.UserRepository
}

// NewUserService is a factory function for
// initializing a UserService with its repository layer dependencies
func NewUserService(c *USConfig) domain.UserService {
	return &UserService{
		UserRepository: c.UserRepository,
	}
}

// Get will retrieve a user with the specified uuid
func (s UserService) Get(ctx context.Context, uid uuid.UUID) (*domain.User, error) {
	u, err := s.UserRepository.FindByID(ctx, uid)

	return u, err
}

// SignUp reaches a UserRepository to verify the
// email address is available and signs up the user if this is the case
func (s UserService) SignUp(ctx context.Context, u *domain.User) error {
	panic("implement me")
}
