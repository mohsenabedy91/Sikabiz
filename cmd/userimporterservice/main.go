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
	"time"
)

const WorkerCount = 10
const maxRetries = 3
const retryDelay = 2 * time.Second

var timestampBackupFile = time.Now().Format("20060102_150405")

func main() {
	configProvider := &config.Config{}
	conf := configProvider.GetConfig()
	log := logger.NewLogger("User Importer Consumer", conf.Log)

	_, filename, _, _ := runtime.Caller(0)
	sourceURL := path.Join(path.Dir(filename), "/../../users_data.json")

	file, fileErr := os.Open(sourceURL)
	if fileErr != nil {
		log.Fatal(logger.Cache, logger.Startup, fmt.Sprintf("Error opening file: %v", fileErr), nil)
	}
	defer func(file *os.File) {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {

		}
	}(file)

	var users []domain.User
	if decoderErr := json.NewDecoder(file).Decode(&users); decoderErr != nil {
		log.Error(logger.Internal, logger.File, fmt.Sprintf("Error parsing user: %v", decoderErr), nil)
		return
	}

	semaphore := make(chan struct{}, WorkerCount)
	var wg sync.WaitGroup

	queue, queueErr := setup.InitializeQueue(log, conf)
	if queueErr != nil {
		return
	}
	defer queue.Driver.Close()

	ctx := context.Background()
	db, databaseErr := setup.InitializeDatabase(ctx, log, conf)
	if databaseErr != nil {
		log.Fatal(logger.Database, logger.Startup, databaseErr.Error(), nil)
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
					handleFailedPublish(u, saveUserEvent, log)
					return
				}

				if createErr := userService.Create(uow, &user); createErr != nil {
					if rollbackErr := uow.Rollback(); rollbackErr != nil {
						handleFailedPublish(u, saveUserEvent, log)
						return
					}
					return
				}

				if commitErr := uow.Commit(); commitErr != nil {
					handleFailedPublish(u, saveUserEvent, log)
					return
				}
				log.Info(logger.Database, logger.DatabaseInsert, "The user has been inserted successfully!", nil)
			}(user)
		default:
			handleFailedPublish(user, saveUserEvent, log)
		}
	}

	wg.Wait()
}

func retry(attempts int, delay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("failed after %d attempts", attempts)
}

func handleFailedPublish(user domain.User, saveUserEvent *event.SaveUser, log logger.Logger) {
	if err := retry(maxRetries, retryDelay, func() error {
		return saveUserEvent.Publish(user)
	}); err != nil {
		log.Error(logger.Queue, logger.RabbitMQPublish, fmt.Sprintf("Failed to publish user: %v. Error: %v", user.ID, err), nil)

		if backupErr := saveFailedDataToBackup(user); backupErr != nil {
			log.Error(logger.Internal, logger.File, fmt.Sprintf("Failed to save user to backup: %v. Error: %v", user.ID, backupErr), nil)
		}
	}
}

func saveFailedDataToBackup(user domain.User) error {
	backupFile := fmt.Sprintf("failed_users_%s.json", timestampBackupFile)

	file, err := os.OpenFile(backupFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening backup file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	encoder := json.NewEncoder(file)
	if encodeErr := encoder.Encode(user); encodeErr != nil {
		return fmt.Errorf("error writing to backup file: %w", encodeErr)
	}
	return nil
}
