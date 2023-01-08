package siakadu

import (
	"context"
	"fmt"
	"time"

	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

type siakadu struct {
	repo         model.SiakaduRepository
	workerClient model.WorkerClient
}

// NewSiakaduScraper create a new siakadu scraper
func NewSiakaduScraper(repo model.SiakaduRepository, workerClient model.WorkerClient) model.SiakaduScraper {
	return &siakadu{
		repo,
		workerClient,
	}
}

func (s *siakadu) Run() {
	// 9999999999 is current max npm

	ctx := context.Background()

	lastNPM, err := s.repo.GetLastNPMFromCache(ctx)
	switch err {
	default:
		logrus.WithError(err).Error("failed to get last npm from cache")
		logrus.Info("using first NPM from configured NPM valu: ", config.SiakadScraperNPMStartAt())
		lastNPM = config.SiakadScraperNPMStartAt()
	case nil:
		logrus.Info("using NPM from cache: ", lastNPM)
	}

	for i := lastNPM; i <= config.SiakadScraperNPMFinishAt(); i++ {
		logger := logrus.WithFields(logrus.Fields{
			"index":   i,
			"seconds": config.SiakadScrapingDelaySeconds(),
		})
		npm := fmt.Sprintf("%d", i)

		if err := s.workerClient.RegisterSiakadProfilePictureTask(context.Background(), npm); err != nil {
			logger.WithError(err).Error("failed to register siakad profile task")
		}

		logger.Info("success registering NPM: ", npm)

		if i%config.SiakadScrapingDelayIndex() != 0 {
			continue
		}

		logger.Info("setting last NPM to cache: ", i)

		if err := s.repo.SetLastNPMToCache(ctx, i); err != nil {
			logger.WithError(err).Error("failed to set last NPM to cache")
		}

		logger.Info("delaying siakad scraping request")

		time.Sleep(config.SiakadScrapingDelaySeconds())

		logger.Info("delay finish")
	}

}
