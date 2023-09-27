package domain

import (
	"context"
	"github.com/google/uuid"
)

// User defines domain model and its json and db representations
type User struct {
	UID      uuid.UUID `json:"uid" db:"uid"`
	Email    string    `json:"email" db:"email"`
	Password string    `json:"-" db:"password"`
	Name     string    `json:"name" db:"name"`
	ImageURL string    `json:"image_url" db:"image_url"`
	Website  string    `json:"website" db:"website"`
}

// UserService defines methods the handler layer expects
// any service it interacts with to implement
type UserService interface {
	Get(ctx context.Context, uid uuid.UUID) (*User, error)
	SignUp(ctx context.Context, u *User) error
	SignIn(ctx context.Context, u *User) error
}

// UserRepository defines methods the service layer expects
// any repository it interacts with to implement
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, uid uuid.UUID) (*User, error)
	Create(ctx context.Context, u *User) error
}
