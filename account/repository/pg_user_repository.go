package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// pgUserRepository is data/repository implementation
// of service layer UserRepository
type pgUserRepository struct {
	DB *sqlx.DB
}

// NewUserRepository is a factory for initializing User Repositories
func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &pgUserRepository{
		DB: db,
	}
}

// FindByID fetches user by id
func (r pgUserRepository) FindByID(ctx context.Context, uid uuid.UUID) (*domain.User, error) {
	user := &domain.User{}

	query := "SELECT * FROM users WHERE uid=$1"

	// we need to actually check errors as it could be something other than not found
	if err := r.DB.GetContext(ctx, user, query, uid); err != nil {
		return user, apperrors.NewNotFound("uid", uid.String())
	}

	return user, nil
}

// Create reaches out to database SQLX api
func (r pgUserRepository) Create(ctx context.Context, u *domain.User) error {
	l := logger.Get()

	query := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING *"

	if err := r.DB.GetContext(ctx, u, query, u.Email, u.Password); err != nil {
		// check unique constraint
		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			l.Info("Could not create a user",
				zap.String("email", u.Email),
				zap.String("reason", err.Code.Name()),
			)
			return apperrors.NewConflict("email", u.Email)
		}
		l.Info("Could not create a user",
			zap.String("email", u.Email),
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	return nil
}

// FindByEmail retrieves user row by email address
func (r pgUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}

	query := "SELECT * FROM users WHERE email=$1"

	// we need to actually check errors as it could be something other than not found
	if err := r.DB.GetContext(ctx, user, query, email); err != nil {
		return user, apperrors.NewNotFound("email", email)
	}

	return user, nil
}

func (r pgUserRepository) Update(ctx context.Context, u *domain.User) error {
	l := logger.Get()

	query := `
		UPDATE users
		SET name = :name, email = :email, website = :website
		WHERE uid = :uid
		RETURNING *;
	`

	nstmt, err := r.DB.PrepareNamedContext(ctx, query)
	if err != nil {
		l.Error("Unable to prepare user update query",
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	if err = nstmt.GetContext(ctx, u, u); err != nil {
		l.Error("Unable to update user",
			zap.Any("uid", u),
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	return nil
}
