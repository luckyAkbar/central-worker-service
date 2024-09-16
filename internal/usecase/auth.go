// Package usecase contains all usecase implementation
package usecase

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
)

const (
	_headerAuthorization string = "Authorization"
	_authScheme          string = "Bearer"
)

type authUsecase struct {
	sessionRepo model.SessionRepository
	userRepo    model.UserRepository
}

// NewAuthUsecase create a new auth usecase
func NewAuthUsecase(sessionRepo model.SessionRepository, userRepo model.UserRepository) model.AuthUsecase {
	return &authUsecase{
		sessionRepo,
		userRepo,
	}
}

// AuthMiddleware with authorized and setting the user to context and passed to the next handler
// if rejectUnauthRequest is true, this will respond with 401 if the token is not present
// if false and token is not present, will pass to the next handler
// if you set true, you don't need to check the user in context in the next handler
func (u *authUsecase) AuthMiddleware(rejectUnauthRequest bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := _getTokenFromHeader(c.Request())
			if token == "" && rejectUnauthRequest {
				return c.NoContent(http.StatusUnauthorized)
			}

			if token == "" {
				return next(c)
			}

			ctx := c.Request().Context()

			session, err := u.sessionRepo.FindByAccessToken(ctx, token)
			if err != nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			if session.IsAccessTokenExpired() {
				return c.NoContent(http.StatusUnauthorized)
			}

			ctx = setUserToCtx(ctx, model.AuthUser{
				ID:        session.UserID,
				SessionID: session.ID,
			})

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func (u *authUsecase) Login(ctx context.Context, input *model.LoginInput) (*model.Session, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	logger.Info("start login usecase process")

	if err := input.Validate(); err != nil {
		logger.Info("input login invalid")
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         MsgInvalidInput,
		}
	}

	user, err := u.userRepo.FindByEmail(ctx, input.Email)
	switch err {
	default:
		logger.Info("database failed to find user by email: ", err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		break
	}

	hash := helper.CreateHashSHA512([]byte(input.Password))

	logger.Info("hash: ", hash)

	if hash != user.Password {
		logger.Info("login failed because of password mismatch")
		return nil, model.UsecaseError{
			UnderlyingError: ErrForbidden,
			Message:         MsgForbidden,
		}
	}

	session := &model.Session{
		ID:                    helper.GenerateID(),
		UserID:                user.ID,
		CreatedAt:             time.Now().UTC(),
		AccessTokenExpiredAt:  time.Now().Add(config.AccessTokenExpiryHour()).UTC(),
		RefreshTokenExpiredAt: time.Now().Add(config.RefreshTokenExpiryHour()).UTC(),
		AccessToken:           helper.GenerateToken(config.AccessTokenLength()),
		RefreshToken:          helper.GenerateToken(config.RefreshTokenLength()),
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		logger.Info("failed to create session: ", err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return session, model.NilUsecaseError
}

func _getTokenFromHeader(req *http.Request) string {
	authHeader := strings.Split(req.Header.Get(_headerAuthorization), " ")

	if len(authHeader) != 2 || authHeader[0] != _authScheme {
		return ""
	}

	return strings.TrimSpace(authHeader[1])
}

// SetUserToCtx self explained
func setUserToCtx(ctx context.Context, user model.AuthUser) context.Context {
	return context.WithValue(ctx, model.UserCtxKey, user)
}
