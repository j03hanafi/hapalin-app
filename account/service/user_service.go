package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"mime/multipart"
	"net/url"
	"path"
)

// userService acts as a struct for injecting an implementation of UserRepository
// for use in service methods
type userService struct {
	UserRepository  domain.UserRepository
	ImageRepository domain.ImageRepository
}

// USConfig will hold repositories that will eventually be injected into this
//
//	service layer
type USConfig struct {
	UserRepository  domain.UserRepository
	ImageRepository domain.ImageRepository
}

// NewUserService is a factory function for
// initializing a userService with its repository layer dependencies
func NewUserService(c *USConfig) domain.UserService {
	return &userService{
		UserRepository:  c.UserRepository,
		ImageRepository: c.ImageRepository,
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
	uFetched, err := s.UserRepository.FindByEmail(ctx, u.Email)

	// Will return NotAuthorized to client to omit details of why
	if err != nil {
		return apperrors.NewAuthorization("Invalid email/password combination")
	}

	// verify password - we previously created this method
	match, err := comparePasswords(uFetched.Password, u.Password)
	if err != nil {
		return apperrors.NewInternal()
	}

	if !match {
		return apperrors.NewAuthorization("Invalid email/password combination")
	}

	*u = *uFetched
	return nil
}

func (s userService) UpdateDetails(ctx context.Context, u *domain.User) error {
	// Update user details in repository
	err := s.UserRepository.Update(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func (s userService) SetProfileImage(ctx context.Context, uid uuid.UUID, imageFileHeader *multipart.FileHeader) (*domain.User, error) {
	l := logger.Get()

	user, err := s.UserRepository.FindByID(ctx, uid)

	if err != nil {
		return nil, err
	}

	objName, err := objNameFromURL(user.ImageURL)

	if err != nil {
		return nil, err
	}

	imageFile, err := imageFileHeader.Open()

	if err != nil {
		l.Error("Failed to open image file",
			zap.Error(err),
		)
		return nil, apperrors.NewInternal()
	}

	// upload user's image to ImageRepository
	// Possibly received updated imageURL
	imageURL, err := s.ImageRepository.UpdateProfile(ctx, objName, imageFile)

	if err != nil {
		l.Error("Failed to upload image to cloud storage",
			zap.Error(err),
		)
		return nil, err
	}

	updatedUser, err := s.UserRepository.UpdateImage(ctx, uid, imageURL)

	if err != nil {
		l.Error("Failed to update user profile image URL",
			zap.Error(err),
		)
		return nil, err
	}

	return updatedUser, nil

}

func (s userService) ClearProfileImage(ctx context.Context, uid uuid.UUID) error {
	user, err := s.UserRepository.FindByID(ctx, uid)
	if err != nil {
		return err
	}

	if user.ImageURL == "" {
		return nil
	}

	objName, err := objNameFromURL(user.ImageURL)
	if err != nil {
		return err
	}

	err = s.ImageRepository.DeleteProfile(ctx, objName)
	if err != nil {
		return err
	}

	_, err = s.UserRepository.UpdateImage(ctx, uid, "")
	if err != nil {
		return err
	}

	return nil
}

func objNameFromURL(imageURL string) (string, error) {
	l := logger.Get()

	// if user doesn't have a profile image, create one
	// otherwise, extract last part of URL to get cloud storage object name
	if imageURL == "" {
		objID, _ := uuid.NewRandom()
		return objID.String(), nil
	}

	// split off last part of URL, which is the image's storage object ID
	urlPath, err := url.Parse(imageURL)

	if err != nil {
		l.Error("Failed to parse image URL",
			zap.Error(err),
		)
		return "", apperrors.NewInternal()
	}

	// get "path" of URL, which is everything after the domain name
	// then get "base", which is the last part of the path
	return path.Base(urlPath.Path), nil
}
