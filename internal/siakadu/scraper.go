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
	for i := config.SiakadScraperNPMStartAt(); i <= config.SiakadScraperNPMFinishAt(); i++ {
		npm := fmt.Sprintf("%d", i)
		if err := s.workerClient.RegisterSiakadProfilePictureTask(context.Background(), npm); err != nil {
			logrus.Error(err)
		}

		logrus.Info("success registering NPM: ", npm)

		if i%config.SiakadScrapingDelayIndex() == 0 {
			logger := logrus.WithFields(logrus.Fields{
				"index":   i,
				"seconds": config.SiakadScrapingDelaySeconds(),
			})
			logger.Info("delaying siakad scraping request")

			time.Sleep(config.SiakadScrapingDelaySeconds())

			logger.Info("delay finish")
		}
	}

}
