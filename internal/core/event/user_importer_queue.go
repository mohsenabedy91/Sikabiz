package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/meesagebroker"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres/userrepository"
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
	"github.com/mohsenabedy91/Sikabiz/internal/core/port"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
)

type SaveUser struct {
	queue       *messagebroker.Queue
	log         logger.Logger
	db          *sql.DB
	userService port.UserService
}

var saveUserInstance *SaveUser

const DelaySaveUserSeconds int64 = 0
const SaveUserName = "save_user_queue"

func NewSaveUser(queue *messagebroker.Queue, log logger.Logger, db *sql.DB, userService port.UserService) *SaveUser {
	if saveUserInstance == nil {
		saveUserInstance = &SaveUser{
			queue:       queue,
			log:         log,
			db:          db,
			userService: userService,
		}
	}

	return saveUserInstance
}

func (r *SaveUser) Name() string {
	return SaveUserName
}

func (r *SaveUser) Publish(message interface{}) error {
	if err := r.queue.Driver.Produce(r.Name(), message, DelaySaveUserSeconds); err != nil {
		return err
	}
	r.queue.Log.Info(
		logger.Queue,
		logger.RabbitMQPublish,
		fmt.Sprintf("published successfully to queue: %s", message),
		nil,
	)

	return nil
}

func (r *SaveUser) Consume(message []byte) error {
	extra := map[logger.ExtraKey]interface{}{
		logger.Body: string(message),
	}
	var user domain.User
	if err := json.Unmarshal(message, &user); err != nil {
		r.queue.Log.Error(logger.Queue, logger.RabbitMQConsume, fmt.Sprintf("Error unmarshalling message, error: %v", err), extra)
		return err
	}

	uowFactory := func() port.UserUnitOfWork {
		return userrepository.NewUnitOfWork(r.log, r.db)
	}
	uow := uowFactory()

	if err := uow.BeginTx(context.Background()); err != nil {
		return err
	}

	if err := r.userService.Create(uow, &user); err != nil {
		if rollbackErr := uow.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	if commitErr := uow.Commit(); commitErr != nil {
		return commitErr
	}

	r.queue.Log.Info(logger.Database, logger.DatabaseInsert, "The message has been consumed successfully!", nil)

	return nil
}

func (r *SaveUser) Register() {
	go func() {
		if err := r.queue.Driver.RegisterConsumer(r.Name(), r.Consume); err != nil {
			r.queue.Log.Error(
				logger.Queue,
				logger.RabbitMQRegisterConsumer,
				fmt.Sprintf("Error on registering consumer, error: %v", err),
				map[logger.ExtraKey]interface{}{
					logger.QueueName: r.Name(),
				},
			)
		}
	}()
}
