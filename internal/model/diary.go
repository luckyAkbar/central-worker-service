package model

import (
	"context"
	"errors"
	"strings"
	"time"
)

// DiarySource is the list of diary source
type DiarySource string

// list of available diary source
var (
	DiarySourceTelegram DiarySource = "telegram"
)

// ParseStringToDiarySource is a function to parse string to DiarySource. if source is not defined
// equally on DiarySource, it will return error
func ParseStringToDiarySource(source string) (DiarySource, error) {
	source = strings.ToLower(source)
	switch source {
	case "telegram":
		return DiarySourceTelegram, nil
	default:
		return "", errors.New("invalid diary source: " + source)
	}
}

// CreateDiaryInput is an input for creating diary
type CreateDiaryInput struct {
	OwnerID   string    `json:"owner_id" validate:"required"`
	Note      string    `json:"note" validate:"required"`
	TimeZone  string    `json:"time_zone" validate:"required"`
	Source    string    `json:"source" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

// Validate validate input
func (cdi *CreateDiaryInput) Validate() error {
	return validator.Struct(cdi)
}

// Diary is a model for diary
type Diary struct {
	ID        string      `json:"id"`
	OwnerID   string      `json:"owner_id"`
	Note      string      `json:"note"`
	CreatedAt time.Time   `json:"created_at"`
	TimeZone  string      `json:"time_zone"`
	Source    DiarySource `json:"source"`
}

// DiaryUsecase is an interface for usecase layer for diary
type DiaryUsecase interface {
	Create(ctx context.Context, input *CreateDiaryInput) (*Diary, UsecaseError)
	GetDiaryByID(ctx context.Context, diaryID, ownerID string) (*Diary, UsecaseError)
	GetDiariesByWrittenDateRange(ctx context.Context, start, end time.Time, ownerID string) ([]Diary, UsecaseError)
	DeleteByID(ctx context.Context, diaryID, ownerID string) UsecaseError
}

// DiaryRepository is an interface for repository layer for diary
type DiaryRepository interface {
	Create(ctx context.Context, diary *Diary) error
	FindDiaryByIDAndOwnerID(ctx context.Context, diaryID, ownerID string) (*Diary, error)
	GetDiariesByWrittenDateRange(ctx context.Context, start, end time.Time, ownerID string) ([]Diary, error)
	DeleteByID(ctx context.Context, diaryID, ownerID string) error
}
