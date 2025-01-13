//go:build !test

package setup

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/meesagebroker"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres"
	"github.com/mohsenabedy91/Sikabiz/internal/core/config"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
)

func InitializeDatabase(ctx context.Context, log logger.Logger, conf config.Config) (*sql.DB, error) {
	if err := postgres.InitClient(ctx, log, conf); err != nil {
		return nil, err
	}
	return postgres.Get(), nil
}

func InitializeQueue(log logger.Logger, conf config.Config) (*messagebroker.Queue, error) {
	queue := messagebroker.NewQueue(log, conf)

	driver, err := messagebroker.NewRabbitMQ(conf.RabbitMQ.URL, log)
	if err != nil {
		log.Fatal(logger.Queue, logger.Startup, fmt.Sprintf("Failed to setup queue, error: %v", err), nil)
		return nil, err
	}

	queue.Driver = driver

	return queue, nil
}
