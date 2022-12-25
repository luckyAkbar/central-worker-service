package model

import (
	"context"

	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	// UserCtxKey context key
	UserCtxKey contextKey = "github.com/luckyAkbar/central-worker-service/model.AuthUser"
)

// AuthUser represents user loaded in context
type AuthUser struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
}

// GetUserFromCtx get user from context
func GetUserFromCtx(ctx context.Context) *AuthUser {
	user, ok := ctx.Value(UserCtxKey).(AuthUser)
	if !ok {
		return nil
	}
	return &user
}

// LoginInput login input
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Validate validate struct
func (li *LoginInput) Validate() error {
	return validator.Struct(li)
}

// AuthUsecase usecase auth
type AuthUsecase interface {
	AuthMiddleware(rejectUnauthReq bool) echo.MiddlewareFunc
	Login(ctx context.Context, input *LoginInput) (*Session, UsecaseError)
}
