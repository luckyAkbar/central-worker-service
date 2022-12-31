package config

import (
	"fmt"
	"time"

	"github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("../../../")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Error(err)
		panic("failed to read config file")
	}
}

// Env returns server env
func Env() string {
	return viper.GetString("server.env")
}

// LogLevel returns server log level
func LogLevel() string {
	return viper.GetString("server.log.level")
}

// PostgresDSN returns postgres DSN
func PostgresDSN() string {
	host := viper.GetString("postgres.host")
	db := viper.GetString("postgres.db")
	user := viper.GetString("postgres.user")
	pw := viper.GetString("postgres.pw")
	port := viper.GetString("postgres.port")
	sslMode := viper.GetString("postgres.ssl_mode")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, pw, db, port, sslMode)
}

// SentryDSN returns sentry dsn
func SentryDSN() string {
	return viper.GetString("server.log.sentry.dsn")
}

// RedisAddr redis address
func RedisAddr() string {
	return viper.GetString("redis.addr")
}

// RedisPassword redis password
func RedisPassword() string {
	return viper.GetString("redis.password")
}

// RedisCacheDB redis db
func RedisCacheDB() int {
	return viper.GetInt("redis.db")
}

// RedisMinIdleConn min idle
func RedisMinIdleConn() int {
	return viper.GetInt("redis.min")
}

// RedisMaxIdleConn max idle
func RedisMaxIdleConn() int {
	return viper.GetInt("redis.max")
}

// ServerPort return server port defined on config file. default to 8080
func ServerPort() string {
	cfg := viper.GetString("server.port")
	if cfg == "" {
		return ":8080"
	}

	return fmt.Sprintf(":%s", cfg)
}

// WorkerConcurrency return worker concurrency defined on config file. default to 10
func WorkerConcurrency() int {
	cfg := viper.GetInt("worker.concurrency")

	if cfg == 0 {
		return 10
	}

	return cfg
}

// MailingTaskMaxRetry max retry for mailing task. default to 5
func MailingTaskMaxRetry() int {
	cfg := viper.GetInt("worker.task.mailing.max_retry")

	if cfg == 0 {
		return 5
	}

	return cfg
}

// MailingTaskTimeoutSeconds timeout for mailing task. default to 5s
func MailingTaskTimeoutSeconds() time.Duration {
	cfg := viper.GetInt("worker.task.mailing.timeout_seconds")

	if cfg == 0 {
		return time.Second * 5
	}

	return time.Second * time.Duration(cfg)
}

// MailUpdatingTaskMaxRetry max retry for mail updating task
func MailUpdatingTaskMaxRetry() int {
	cfg := viper.GetInt("worker.task.mail_updating.max_retry")

	if cfg == 0 {
		return 5
	}

	return cfg
}

// MailUpdatingTaskTimeoutSeconds timeout for mail updating task
func MailUpdatingTaskTimeoutSeconds() time.Duration {
	cfg := viper.GetInt("worker.task.mail_updating.timeout_seconds")

	if cfg == 0 {
		return time.Second * 5
	}

	return time.Second * time.Duration(cfg)
}

// UserActivationTaskMaxRetry max retry for mail updating task
func UserActivationTaskMaxRetry() int {
	cfg := viper.GetInt("worker.task.user_activation.max_retry")
	logrus.Info("user activation task max retry: ", cfg)

	if cfg == 0 {
		return 5
	}

	return cfg
}

// UserActivationTaskTimeoutSeconds timeout for mail updating task
func UserActivationTaskTimeoutSeconds() time.Duration {
	cfg := viper.GetInt("worker.task.user_activation.timeout_seconds")
	logrus.Info("user activation task max retry: ", cfg)

	if cfg == 0 {
		return time.Second * 5
	}

	return time.Second * time.Duration(cfg)
}

// ServerSenderName name for email in sending
func ServerSenderName() string {
	return viper.GetString("server.sender.name")
}

// ServerSenderEmail email for email address in sending
func ServerSenderEmail() string {
	return viper.GetString("server.sender.email")
}

// SendinblueAPIKey get API key for send in blue
func SendinblueAPIKey() string {
	return viper.GetString("sendinblue.api_key")
}

// SendInBlueSender generate sendinblue sender using configured sender name and sender email
func SendInBlueSender() *lib.SendSmtpEmailSender {
	return &lib.SendSmtpEmailSender{
		Name:  ServerSenderName(),
		Email: ServerSenderEmail(),
	}
}

// SendInBlueIsActivated is activated sendinblue
func SendInBlueIsActivated() bool {
	return viper.GetBool("sendinblue.is_activated")
}

// MailgunIsActivated is activated mailgun
func MailgunIsActivated() bool {
	return viper.GetBool("mailgun.is_activated")
}

// MailgunDomain mailgun domain
func MailgunDomain() string {
	return viper.GetString("mailgun.domain")
}

// MailgunPrivateAPIKey mailgun private api key
func MailgunPrivateAPIKey() string {
	return viper.GetString("mailgun.private_api_key")
}

// MailgunPublicAPIKey mailgun public api key
func MailgunPublicAPIKey() string {
	return viper.GetString("mailgun.public_api_key")
}

// WorkerBrokerRedisHost redis host for worker task broker
func WorkerBrokerRedisHost() string {
	return viper.GetString("redis.worker_broker_host")
}

// MinUserPasswordLength return minimum length of user password
func MinUserPasswordLength() int {
	cfg := viper.GetInt("server.user.min_password_length")

	if cfg == 0 {
		return 8
	}

	return cfg
}

// UserActivationBaseURL return user activation url
func UserActivationBaseURL() string {
	return viper.GetString("server.user.activation.base_url")
}

// NewRelicLisence new relic lisence
func NewRelicLisence() string {
	return viper.GetString("newrelic.lisence")
}

// NewRelicLoggingLogForwarding nr log forwarding
func NewRelicLoggingLogForwarding() bool {
	return viper.GetBool("newrelic.logging.log_forwarding_enabled")
}

// NewRelicLoggingAppLogEnabled app log enabled
func NewRelicLoggingAppLogEnabled() bool {
	return viper.GetBool("newrelic.logging.app_log_enabled")
}

// NewRelicLoggingLogDecorationEnabled log decoration enabled
func NewRelicLoggingLogDecorationEnabled() bool {
	return viper.GetBool("newrelic.logging.log_decoration_enabled")
}

// NewRelicServerAppName server app name
func NewRelicServerAppName() string {
	return viper.GetString("newrelic.server.app_name")
}

// NewRelicWorkerAppName worker app name
func NewRelicWorkerAppName() string {
	return viper.GetString("newrelic.worker.app_name")
}

// AccessTokenExpiryHour expiry hour for access token
func AccessTokenExpiryHour() time.Duration {
	cfg := viper.GetInt("server.auth.session.access_token_expiry_hour")
	if cfg == 0 {
		cfg = 8
	}

	return time.Hour * time.Duration(cfg)
}

// RefreshTokenExpiryHour expiry hour for refresh token
func RefreshTokenExpiryHour() time.Duration {
	cfg := viper.GetInt("server.auth.session.refresh_token_expiry_hour")
	if cfg == 0 {
		cfg = 8
	}

	return time.Hour * time.Duration(cfg)
}

// AccessTokenLength length of the access token
func AccessTokenLength() int {
	cfg := viper.GetInt("server.auth.session.access_token_length")
	if cfg == 0 {
		return 8
	}

	return cfg
}

// RefreshTokenLength length of the refresh token
func RefreshTokenLength() int {
	cfg := viper.GetInt("server.auth.session.refresh_token_length")
	if cfg == 0 {
		return 8
	}

	return cfg
}

// TelegramBotToken get telegram bot token
func TelegramBotToken() string {
	cfg := viper.GetString("telegram_bot.token")
	if cfg == "" {
		panic("telegram bot token is not exists")
	}

	return cfg
}

// TelegramBotUseTestEnv return whether to use test env by gotgbot module
func TelegramBotUseTestEnv() bool {
	return viper.GetBool("telegram_bot.environment.use_test_env")
}

// TelegramBotTimeoutDuration timeout duration for telegram bot
func TelegramBotTimeoutDuration() time.Duration {
	cfg := viper.GetInt("telegram_bot.environment.timeout")
	if cfg == 0 {
		return time.Second * 10
	}

	return time.Second * time.Duration(cfg)
}

// TelegramBotTimeout timeout for telegram bot
func TelegramBotTimeout() int64 {
	cfg := viper.GetInt64("telegram_bot.environment.timeout")
	if cfg == 0 {
		return 10
	}

	return cfg
}

// TelegramBotDropPendingUpdate decide wheter to drop pending update from telegram bot
func TelegramBotDropPendingUpdate() bool {
	return viper.GetBool("telegram_bot.environment.drop_pending_update")
}

// ImageMediaAllowedTypes allowed mimetype for image media
func ImageMediaAllowedTypes() []string {
	return viper.GetStringSlice("server.media.image.allowed_types")
}

// ImageMediaMaxSizeBytes max size in bytes for image media
func ImageMediaMaxSizeBytes() int64 {
	return viper.GetInt64("server.media.image.max_size_bytes")
}

// ImageMediaTokenLength token length for image media
func ImageMediaTokenLength() int {
	return viper.GetInt("server.media.image.token.length")
}

// ImageMediaLocalStorage local storage for image media
func ImageMediaLocalStorage() string {
	cfg := viper.GetString("server.media.image.local_storage")
	if cfg == "" {
		panic("config for server.media.image.local_storage is not defined")
	}

	return cfg
}
