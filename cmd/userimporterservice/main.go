package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mohsenabedy91/Sikabiz/cmd/setup"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres/userrepository"
	"github.com/mohsenabedy91/Sikabiz/internal/core/config"
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
	"github.com/mohsenabedy91/Sikabiz/internal/core/event"
	"github.com/mohsenabedy91/Sikabiz/internal/core/port"
	"github.com/mohsenabedy91/Sikabiz/internal/core/service/userservice"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	"os"
	"path"
	"runtime"
	"sync"
)

const WorkerCount = 10

func main() {
	configProvider := &config.Config{}
	conf := configProvider.GetConfig()
	log := logger.NewLogger("User Importer Consumer", conf.Log)

	_, filename, _, _ := runtime.Caller(0)
	sourceURL := path.Join(path.Dir(filename), "/../../users_data.json")

	file, err := os.Open(sourceURL)
	if err != nil {
		log.Fatal(logger.Cache, logger.Startup, fmt.Sprintf("Error opening file: %v", err), nil)
	}
	defer func(file *os.File) {
		fileErr := file.Close()
		if fileErr != nil {

		}
	}(file)

	var users []domain.User
	if decoderErr := json.NewDecoder(file).Decode(&users); err != nil {
		log.Error(logger.Internal, logger.File, fmt.Sprintf("Error parsing user: %v", decoderErr), nil)
		return
	}

	semaphore := make(chan struct{}, WorkerCount)
	var wg sync.WaitGroup

	queue, err := setup.InitializeQueue(log, conf)
	if err != nil {
		return
	}
	defer queue.Driver.Close()

	ctx := context.Background()
	db, err := setup.InitializeDatabase(ctx, log, conf)
	if err != nil {
		log.Fatal(logger.Database, logger.Startup, err.Error(), nil)
		return
	}

	userService := userservice.New(log)
	saveUserEvent := event.NewSaveUser(queue, log, db, userService)

	uowFactory := func() port.UserUnitOfWork {
		return userrepository.NewUnitOfWork(log, db)
	}

	for _, user := range users {
		select {
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func(u domain.User) {
				defer func() {
					wg.Done()
					<-semaphore
				}()

				uow := uowFactory()
				if txErr := uow.BeginTx(ctx); txErr != nil {
					return
				}

				if createErr := userService.Create(uow, &user); createErr != nil {
					if rollbackErr := uow.Rollback(); rollbackErr != nil {
						return
					}
					return
				}

				if commitErr := uow.Commit(); commitErr != nil {
					return
				}
				log.Info(logger.Database, logger.DatabaseInsert, "The user has been inserted successfully!", nil)
			}(user)
		default:
			saveUserEvent.Publish(user)
		}
	}

	wg.Wait()
}
