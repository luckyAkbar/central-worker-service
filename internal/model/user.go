package model

import (
	"context"
	"fmt"
	"time"

	"github.com/luckyAkbar/central-worker-service/internal/config"
	"gopkg.in/guregu/null.v4"
)

// RegisterUserInput input for register user
type RegisterUserInput struct {
	Email                string `json:"email" validate:"required,email"`
	Username             string `json:"username" validate:"required"`
	Password             string `json:"password" validate:"required,eqfield=PasswordConfirmation"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}

// Validate validate struct, also check if the user password is less than configured minimum length
func (rui *RegisterUserInput) Validate() error {
	if len(rui.Password) < config.MinUserPasswordLength() {
		return fmt.Errorf("password length must be at least %d", config.MinUserPasswordLength())
	}

	return validator.Struct(rui)
}

// User represent user on database
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt null.Time `json:"deleted_at"`
	IsActive  bool      `json:"is_active"`
}

// GenerateActivationSignatureInput generate signature with format: id-username-email
func (u *User) GenerateActivationSignatureInput() string {
	return fmt.Sprintf("%s-%s-%s", u.ID, u.Username, u.Email)
}

// UserUsecase usecase for user
type UserUsecase interface {
	Register(ctx context.Context, input *RegisterUserInput) (*User, UsecaseError)
}

// UserRepository repository for user
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}
