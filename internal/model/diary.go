package model

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
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

// DiaryListForFrontend is component for diary accordion
type DiaryListForFrontend struct {
	DiaryAccordionHeader string
	DiaryAccordionBody   string
}

// DiaryFrontendTemplateData is the data structure representing the template data to render
// HTML page for diary
type DiaryFrontendTemplateData struct {
	LeadText  string
	DiaryList []DiaryListForFrontend
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

func (d Diary) LenNoteChars() int64 {
	return int64(utf8.RuneCountInString(d.Note))
}

type DiaryList []Diary

func (dl DiaryList) ToDiaryFrontendTemplateData(leadText string) DiaryFrontendTemplateData {
	dlff := []DiaryListForFrontend{}
	for _, diary := range dl {
		dlff = append(dlff, DiaryListForFrontend{
			DiaryAccordionHeader: diary.CreatedAt.Format(time.DateOnly),
			DiaryAccordionBody:   diary.Note,
		})
	}

	return DiaryFrontendTemplateData{
		LeadText:  leadText,
		DiaryList: dlff,
	}
}

// DiaryUsecase is an interface for usecase layer for diary
type DiaryUsecase interface {
	Create(ctx context.Context, input *CreateDiaryInput) (*Diary, UsecaseError)
	GetDiaryByID(ctx context.Context, diaryID, ownerID string) (*Diary, UsecaseError)
	GetDiariesByWrittenDateRange(ctx context.Context, start, end time.Time, ownerID string) ([]Diary, UsecaseError)
	DeleteByID(ctx context.Context, diaryID, ownerID string) UsecaseError
	PrepareRenderDiariesOnFrontend(ctx context.Context, cmd []string, diaries DiaryList) (string, error)
	GetDiariesOnFrontendRenderData(ctx context.Context, key string) (*DiaryFrontendTemplateData, error)
}

// DiaryRepository is an interface for repository layer for diary
type DiaryRepository interface {
	Create(ctx context.Context, diary *Diary) error
	FindDiaryByIDAndOwnerID(ctx context.Context, diaryID, ownerID string) (*Diary, error)
	GetDiariesByWrittenDateRange(ctx context.Context, start, end time.Time, ownerID string) ([]Diary, error)
	DeleteByID(ctx context.Context, diaryID, ownerID string) error
	SetFrontendDiaryDataToCache(ctx context.Context, cacheKey string, exp time.Duration, data DiaryFrontendTemplateData) error
	GetFrontendDiaryDataFromCache(ctx context.Context, key string) (*DiaryFrontendTemplateData, error)
}
