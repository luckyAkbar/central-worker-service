package model

import (
	"context"
	"fmt"
)

// GagMemeType is a type that represent the type meme media
type GagMemeType string

// list meme media type
var (
	GagMemeTypeImage GagMemeType = "image"
	GagMemeTypeVideo GagMemeType = "video"
)

// GagMeme is a model that represent the gag meme
type GagMeme struct {
	ID          string      `json:"id" gorm:"column:id"`
	OriginalURL string      `json:"original_url" gorm:"column:originalUrl"`
	Type        GagMemeType `json:"type" gorm:"column:type"`
	MediaURL    string      `json:"media_url" gorm:"column:mediaUrl"`
	Title       string      `json:"title" gorm:"column:title"`
}

// TableName override gorm default naming convention
func (gm *GagMeme) TableName() string {
	return "GagMemes"
}

// GenerateCaptionForSubscription will generate caption for gag meme subscription
// include html formatting
func (gm *GagMeme) GenerateCaptionForSubscription() string {
	return fmt.Sprintf(`Meme Subscription
Title: <strong>%s</strong>
Original Post: <strong>%s</strong>
`, gm.Title, gm.OriginalURL)
}

// GagMemeUsecase is a usecase that represent the gag meme usecase
type GagMemeUsecase interface {
	GetRandomGagMeme(ctx context.Context) (*GagMeme, error)
}

// GagMemeRepository is a repository that represent the gag meme repository
type GagMemeRepository interface {
	FindRandomGagMeme(ctx context.Context) (*GagMeme, error)
}
