package config

import (
	"fmt"

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

func Env() string {
	return viper.GetString("server.env")
}

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
