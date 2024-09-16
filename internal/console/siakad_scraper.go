package console

import (
	"github.com/go-redis/redis/v9"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
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

	db.InitializePostgresConn()

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr(),
		Password:     config.RedisPassword(),
		DB:           config.RedisCacheDB(),
		MinIdleConns: config.RedisMinIdleConn(),
		MaxIdleConns: config.RedisMaxIdleConn(),
	})
	cacher := db.NewCacher(redisClient)

	repo := repository.NewSiakadRepository(db.PostgresDB, cacher)

	scraper := siakadu.NewSiakaduScraper(repo, workerClient)

	scraper.Run()
}
