package console

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/delivery/rest"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/middleware"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/luckyAkbar/central-worker-service/internal/worker"
	nrEcho "github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	nrLogrus "github.com/newrelic/go-agent/v3/integrations/nrlogrus"
	nr "github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the server",
	Run:   server,
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

func server(c *cobra.Command, args []string) {
	db.InitializePostgresConn()
	sqlDB, err := db.PostgresDB.DB()
	if err != nil {
		logrus.Fatal("unable to start server. reason: ", err.Error())
	}

	defer helper.WrapCloser(sqlDB.Close)

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr(),
		Password:     config.RedisPassword(),
		DB:           config.RedisCacheDB(),
		MinIdleConns: config.RedisMinIdleConn(),
		MaxIdleConns: config.RedisMaxIdleConn(),
	})
	cacher := db.NewCacher(redisClient)

	mailRepo := repository.NewMailRepository(db.PostgresDB)
	userRepo := repository.NewUserRepository(db.PostgresDB)
	sessionRepo := repository.NewSessionRepository(db.PostgresDB)
	imageRepo := repository.NewImageRepository(db.PostgresDB)
	diaryRepo := repository.NewDiaryRepo(db.PostgresDB, cacher)

	workerClient, err := worker.NewClient(config.WorkerBrokerRedisHost())
	if err != nil {
		logrus.Fatal(err)
	}

	yourlsUtil := helper.NewYourlsUtil(config.YourlsBaseUrl(), config.YourlsSignature(), &http.Client{})

	mailUsecase := usecase.NewMailUsecase(mailRepo, workerClient)
	userUsecase := usecase.NewUserUsecase(userRepo, mailUsecase, workerClient)
	authUsecase := usecase.NewAuthUsecase(sessionRepo, userRepo)
	imageUsecase := usecase.NewImageUsecase(imageRepo)
	diaryUsecase := usecase.NewDiaryUsecase(diaryRepo, yourlsUtil)

	HTTPServer := echo.New()
	HTTPServer.Pre(echoMiddleware.AddTrailingSlash())
	HTTPServer.Use(middleware.RequestID())
	HTTPServer.Use(echoMiddleware.Logger())
	HTTPServer.Use(echoMiddleware.Recover())
	HTTPServer.Use(echoMiddleware.CORS())

	newRelic, nrError := nr.NewApplication(
		nr.ConfigAppName(config.NewRelicServerAppName()),
		nr.ConfigLicense(config.NewRelicLisence()),
		nr.ConfigAppLogForwardingEnabled(config.NewRelicLoggingLogForwarding()),
		nrLogrus.ConfigStandardLogger(),
		nr.ConfigAppLogEnabled(config.NewRelicLoggingAppLogEnabled()),
		nr.ConfigAppLogDecoratingEnabled(config.NewRelicLoggingLogDecorationEnabled()),
	)

	if nrError == nil {
		logrus.Info("adding newrelic echo middleware")
		HTTPServer.Use(nrEcho.Middleware(newRelic))
	}

	apiGroup := HTTPServer.Group("api")
	authGroup := HTTPServer.Group("auth")
	feGroup := HTTPServer.Group("frontend")

	rest.Init(apiGroup, authGroup, feGroup, mailUsecase, userUsecase, authUsecase, imageUsecase, diaryUsecase)

	logrus.Info("starting the server...")
	if err := HTTPServer.Start(config.ServerPort()); err != nil {
		log.Fatal("unable to start server: ", err)
	}
}
