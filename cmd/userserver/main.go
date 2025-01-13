//go:build !test

package main

import (
	"context"
	"fmt"
	"github.com/mohsenabedy91/Sikabiz/cmd/setup"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/handler"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/routes"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres/userrepository"
	"github.com/mohsenabedy91/Sikabiz/internal/core/config"
	"github.com/mohsenabedy91/Sikabiz/internal/core/port"
	"github.com/mohsenabedy91/Sikabiz/internal/core/service/userservice"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	"github.com/mohsenabedy91/Sikabiz/pkg/translation"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @securityDefinitions.apikey AuthBearer
// @in header
// @name Authorization
// @description "Bearer <your-jwt-token>"
func main() {
	configProvider := &config.Config{}
	conf := configProvider.GetConfig()
	log := logger.NewLogger("UserManagement", conf.Log)

	ctx := context.Background()
	defer func() {
		if err := postgres.Close(); err != nil {
			log.Fatal(logger.Database, logger.Startup, err.Error(), nil)
		}
	}()
	postgresDB, err := setup.InitializeDatabase(ctx, log, conf)
	if err != nil {
		log.Fatal(logger.Database, logger.Startup, err.Error(), nil)
		return
	}
	uowFactory := func() port.UserUnitOfWork {
		return userrepository.NewUnitOfWork(log, postgresDB)
	}

	trans := translation.NewTranslation(conf.App)
	trans.GetLocalizer(conf.App.Locale)

	userService := userservice.New(log)

	httpServer := startHTTPServer(log, conf, trans, userService, uowFactory)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	<-signalCh

	log.Info(logger.Internal, logger.Shutdown, "Shutdown Servers ...", nil)

	shutdownHTTPServer(ctx, httpServer, log)
}

func startHTTPServer(
	log logger.Logger,
	conf config.Config,
	trans translation.Translator,
	userService *userservice.UserService,
	uowFactory func() port.UserUnitOfWork,
) *http.Server {

	userHandler := handler.NewUserHandler(trans, userService, uowFactory)

	// Init router
	router, err := routes.NewRouter(log, conf, trans)
	if err != nil {
		log.Fatal(logger.Internal, logger.Startup, err.Error(), nil)
		return nil
	}

	router = router.NewUserRouter(*userHandler)

	listenAddr := fmt.Sprintf("%s:%s", conf.App.HTTPUrl, conf.App.HTTPPort)
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: router.Engine.Handler(),
	}
	log.Info(logger.Internal, logger.Startup, "Starting the HTTP server", map[logger.ExtraKey]interface{}{
		logger.ListeningAddress: httpServer.Addr,
	})

	router.Serve(httpServer)
	return httpServer
}

func shutdownHTTPServer(
	ctx context.Context,
	server *http.Server,
	log logger.Logger,
) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxWithTimeout); err != nil {
		log.Fatal(logger.Internal, logger.Shutdown, fmt.Sprintf("Shutdown Server: %v", err), nil)
	}

	<-ctxWithTimeout.Done()
	log.Info(logger.Internal, logger.Shutdown, "Server exiting", nil)
}
