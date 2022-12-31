package console

import (
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/siakadu"
	"github.com/luckyAkbar/central-worker-service/internal/worker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var siakadScraperCmd = &cobra.Command{
	Use:   "siakad-scraper",
	Short: "run profile photo scraping on siakad server",
	Run:   runSiakadScraper,
}

func init() {
	RootCmd.AddCommand(siakadScraperCmd)
}

func runSiakadScraper(c *cobra.Command, args []string) {
	workerClient, err := worker.NewClient(config.WorkerBrokerRedisHost())
	if err != nil {
		logrus.Fatal(err)
	}

	scraper := siakadu.NewSiakaduScraper(nil, workerClient)

	scraper.Run()
}
