package model

import (
	"context"
	"time"
)

// Session session on db
type Session struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	CreatedAt             time.Time `json:"created_at"`
	AccessTokenExpiredAt  time.Time `json:"access_token_expired_at"`
	RefreshTokenExpiredAt time.Time `json:"refresh_token_expired_at"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
}

// IsAccessTokenExpired check is access token expired at is after now
func (s *Session) IsAccessTokenExpired() bool {
	if s == nil {
		return true
	}

	now := time.Now()
	return now.After(s.AccessTokenExpiredAt)
}

// SessionRepository session repository
type SessionRepository interface {
	Create(ctx context.Context, sess *Session) error
	FindByAccessToken(ctx context.Context, accessToken string) (*Session, error)
}
