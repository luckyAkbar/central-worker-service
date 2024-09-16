package model

import (
	"context"
	"time"
)

const (
	// LastNPMCacheKey cache key to store last NPM crawled by siakadu scraper
	LastNPMCacheKey = "github.com/luckyAkbar/central_service/model/siakadu:last_npm_cache_key"
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
	GetLastNPMFromCache(ctx context.Context) (int, error)
	SetLastNPMToCache(ctx context.Context, npm int) error
}
