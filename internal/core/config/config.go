package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	once   sync.Once
	config Config
)

type Config struct {
	App      App
	DB       DB
	Log      Log
	Swagger  Swagger
	RabbitMQ RabbitMQ
}

type App struct {
	Name               string
	Env                string
	Debug              bool
	Timezone           string
	Locale             string
	PathLocale         string
	GracefullyShutdown time.Duration
	HTTPUrl            string
	HTTPPort           string
}

type Log struct {
	FilePath   string
	Level      string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

type SwaggerInfo struct {
	Title       string
	Description string
	Version     string
}

type Swagger struct {
	Host     string
	Schemes  string
	Info     SwaggerInfo
	Enable   bool
	Username string
	Password string
}

type DBPostgres struct {
	SSLMode            string
	MaxOpenConnections int
	MaxIdleConnections int
	MaxLifetime        time.Duration
	Timezone           string
}

type DB struct {
	Connection string
	Host       string
	Port       string
	Name       string
	Username   string
	Password   string
	Postgres   DBPostgres
}

type RabbitMQ struct {
	URL string
}

type Configuration interface {
	LoadConfig(envPath ...string) (Config, error)
	GetConfig(envPath ...string) Config
}

func (r *Config) LoadConfig(envPath ...string) (Config, error) {
	err := godotenv.Load(envPath...)
	if err != nil {
		return Config{}, fmt.Errorf("error loading .env file: %v", err)
	}

	var app App
	app.Name = os.Getenv("APP_NAME")
	app.Env = os.Getenv("APP_ENV")
	app.Debug = getBoolEnv("APP_DEBUG", false)
	app.Timezone = os.Getenv("APP_TIMEZONE")
	app.Locale = os.Getenv("APP_LOCALE")
	app.PathLocale = os.Getenv("APP_PATH_LOCALE")
	app.GracefullyShutdown = time.Duration(getIntEnv("APP_GRACEFULLY_SHUTDOWN", 5))
	app.HTTPUrl = os.Getenv("HTTP_URL")
	app.HTTPPort = os.Getenv("HTTP_PORT")

	var db DB
	db.Connection = os.Getenv("DB_CONNECTION")
	db.Host = os.Getenv("DB_HOST")
	db.Port = os.Getenv("DB_PORT")
	db.Name = os.Getenv("DB_NAME")
	db.Username = os.Getenv("DB_USERNAME")
	db.Password = os.Getenv("DB_PASSWORD")
	db.Postgres.SSLMode = os.Getenv("DB_POSTGRES_SSL_MODE")
	db.Postgres.MaxOpenConnections = getIntEnv("DB_POSTGRES_MAX_OPEN_CONNECTIONS", 0)
	db.Postgres.MaxIdleConnections = getIntEnv("DB_POSTGRES_MAX_IDLE_CONNECTIONS", 0)
	db.Postgres.MaxLifetime = time.Duration(getIntEnv("DB_POSTGRES_MAX_LIFETIME", 0))
	db.Postgres.Timezone = os.Getenv("DB_POSTGRES_TIMEZONE")

	var log Log
	log.FilePath = os.Getenv("LOG_FILE_PATH")
	log.Level = os.Getenv("LOG_LEVEL")
	log.MaxSize = getIntEnv("LOG_MAX_SIZE", 1)
	log.MaxAge = getIntEnv("LOG_MAX_AGE", 5)
	log.MaxBackups = getIntEnv("LOG_MAX_BACKUPS", 10)

	var swagger Swagger
	swagger.Host = os.Getenv("SWAGGER_HOST")
	swagger.Schemes = os.Getenv("SWAGGER_SCHEMES")
	swagger.Info.Title = os.Getenv("SWAGGER_INFO_TITLE")
	swagger.Info.Description = os.Getenv("SWAGGER_INFO_DESCRIPTION")
	swagger.Info.Version = os.Getenv("SWAGGER_INFO_VERSION")
	swagger.Enable = getBoolEnv("SWAGGER_ENABLE", false)
	swagger.Username = os.Getenv("SWAGGER_USERNAME")
	swagger.Password = os.Getenv("SWAGGER_PASSWORD")

	var rabbitMQ RabbitMQ
	rabbitMQ.URL = os.Getenv("RABBITMQ_URL")

	return Config{
		App:      app,
		DB:       db,
		Log:      log,
		Swagger:  swagger,
		RabbitMQ: rabbitMQ,
	}, nil
}

func getBoolEnv(key string, defaultValue bool) bool {
	val, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return val
}

func getIntEnv(key string, defaultValue int) int {
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return val
}

func (r *Config) GetConfig(envPath ...string) Config {
	once.Do(func() {
		var err error
		config, err = r.LoadConfig(envPath...)
		if err != nil {
			panic(err)
		}
	})
	return config
}
