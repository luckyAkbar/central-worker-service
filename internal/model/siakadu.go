package model

import (
	"context"
	"time"
)

// SiakaduScrapingResult represents the db table
type SiakaduScrapingResult struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Filename  string    `json:"filename"`
	Location  string    `json:"location"`
}

// SiakaduScraper siakadu scraper
type SiakaduScraper interface {
	Run()
}

// SiakaduRepository repository
type SiakaduRepository interface {
	Create(ctx context.Context, result *SiakaduScrapingResult) error
	FindByID(ctx context.Context, id string) (*SiakaduScrapingResult, error)
}
