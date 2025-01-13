//go:build !test

package main

import (
	"context"
	"github.com/mohsenabedy91/Sikabiz/cmd/setup"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/meesagebroker"
	"github.com/mohsenabedy91/Sikabiz/internal/core/config"
	"github.com/mohsenabedy91/Sikabiz/internal/core/event"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configProvider := &config.Config{}
	conf := configProvider.GetConfig()
	log := logger.NewLogger("User Importer Consumer", conf.Log)

	queue, err := setup.InitializeQueue(log, conf)
	if err != nil {
		return
	}
	defer queue.Driver.Close()

	ctx := context.Background()
	postgresDB, err := setup.InitializeDatabase(ctx, log, conf)
	if err != nil {
		log.Fatal(logger.Database, logger.Startup, err.Error(), nil)
		return
	}

	log.Info(logger.Queue, logger.Startup, "Setup queue successfully", nil)

	messagebroker.RegisterEvents(
		event.NewSaveUser(queue, log, postgresDB),
		// add new queues here
		// ...
	)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	<-signalCh

	log.Info(logger.Internal, logger.Shutdown, "Shutdown Server ...", nil)
}
