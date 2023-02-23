package model

import (
	"context"
	"fmt"
)

type GagMemeType string

var (
	GagMemeTypeImage GagMemeType = "image"
	GagMemeTypeVideo GagMemeType = "video"
)

type GagMeme struct {
	ID          string      `json:"id" gorm:"column:id"`
	OriginalURL string      `json:"original_url" gorm:"column:originalUrl"`
	Type        GagMemeType `json:"type" gorm:"column:type"`
	MediaURL    string      `json:"media_url" gorm:"column:mediaUrl"`
	Title       string      `json:"title" gorm:"column:title"`
}

func (gm *GagMeme) TableName() string {
	return "GagMemes"
}

func (gm *GagMeme) GenerateCaptionForSubscription() string {
	return fmt.Sprintf(`Meme Subscription
Title: <strong>%s</strong>
Original Post: <strong>%s</strong>
`, gm.Title, gm.OriginalURL)
}

type GagMemeUsecase interface {
	GetRandomGagMeme(ctx context.Context) (*GagMeme, error)
}

type GagMemeRepository interface {
	FindRandomGagMeme(ctx context.Context) (*GagMeme, error)
}
