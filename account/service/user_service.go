package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
)

// userService acts as a struct for injecting an implementation of UserRepository
// for use in service methods
type userService struct {
	UserRepository domain.UserRepository
}

// USConfig will hold repositories that will eventually be injected into this
//
//	service layer
type USConfig struct {
	UserRepository domain.UserRepository
}

// NewUserService is a factory function for
// initializing a userService with its repository layer dependencies
func NewUserService(c *USConfig) domain.UserService {
	return &userService{
		UserRepository: c.UserRepository,
	}
}

// Get will retrieve a user with the specified uuid
func (s userService) Get(ctx context.Context, uid uuid.UUID) (*domain.User, error) {
	u, err := s.UserRepository.FindByID(ctx, uid)

	return u, err
}

// SignUp reaches a UserRepository to verify the
// email address is available and signs up the user if this is the case
func (s userService) SignUp(ctx context.Context, u *domain.User) error {
	l := logger.Get()

	pw, err := hashPassword(u.Password)
	if err != nil {
		l.Error("error hashing password",
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	u.Password = pw
	if err := s.UserRepository.Create(ctx, u); err != nil {
		return err
	}

	// If we get around to adding events, we'd Publish it here
	// err := s.EventsBroker.PublishUserUpdated(u, true)

	// if err != nil {
	//  return nil, apperrors.NewInternal()
	// }

	return nil
}

// SignIn reaches our to a UserRepository check if the user exists
// and then compares the supplied password with the provided password
// if a valid email/password combo is provided, u will hold all
// available user fields
func (s userService) SignIn(ctx context.Context, u *domain.User) error {
	panic("not implemented")
}
