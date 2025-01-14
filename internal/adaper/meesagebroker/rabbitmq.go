package messagebroker

import (
	"encoding/json"
	"fmt"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

type RabbitMQ struct {
	conn         *amqp.Connection
	log          logger.Logger
	url          string
	notifyClose  chan *amqp.Error
	mu           sync.Mutex
	consumers    map[string]func(message []byte) error
	consumerLock sync.Mutex
}

func NewRabbitMQ(url string, log logger.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	rmq := &RabbitMQ{
		conn:        conn,
		log:         log,
		url:         url,
		notifyClose: conn.NotifyClose(make(chan *amqp.Error)),
		consumers:   make(map[string]func(message []byte) error),
	}

	go rmq.handleReconnect()

	return rmq, nil
}

func (r *RabbitMQ) Close() {
	if err := r.conn.Close(); err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQ, err.Error(), nil)
	}
}

func (r *RabbitMQ) Produce(name string, msg interface{}, delaySeconds int64) error {
	message, err := json.Marshal(msg)
	if err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error marshalling value: %v", err), nil)
		return err
	}

	extra := map[logger.ExtraKey]interface{}{
		logger.QueueName: name,
		logger.Body:      message,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	channel, err := r.conn.Channel()
	if err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error create channel: %v", err), extra)
		return err
	}

	defer func(ch *amqp.Channel) {
		if err = ch.Close(); err != nil {
			r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error closing channel: %v", err), extra)
		}
	}(channel)

	if err = channel.ExchangeDeclare(
		"delayed_exchange",
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp.Table{"x-delayed-type": "direct"},
	); err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error ExchangeDeclare: %v", err), extra)
		return err
	}

	queueDeclare, err := channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error QueueDeclare: %v", err), extra)
		return err
	}

	if err = channel.QueueBind(
		queueDeclare.Name,
		queueDeclare.Name,
		"delayed_exchange",
		false,
		nil,
	); err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error QueueBind: %v", err), extra)
		return err
	}

	if err = channel.Publish(
		"delayed_exchange",
		queueDeclare.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			Headers:      amqp.Table{"x-delay": delaySeconds * 1000},
			DeliveryMode: amqp.Persistent,
		},
	); err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQProduce, fmt.Sprintf("Error Publish message: %v", err), extra)
		return err
	}

	return nil
}

func (r *RabbitMQ) RegisterConsumer(name string, callback func(message []byte) error) error {
	r.consumerLock.Lock()
	defer r.consumerLock.Unlock()

	r.consumers[name] = callback

	return r.setupConsumer(name, callback)
}

func (r *RabbitMQ) setupConsumer(name string, callback func(message []byte) error) error {
	extra := map[logger.ExtraKey]interface{}{
		logger.QueueName: name,
	}

	r.mu.Lock()
	channel, err := r.conn.Channel()
	r.mu.Unlock()
	if err != nil {
		r.log.Error(logger.Queue, logger.RabbitMQRegisterConsumer, fmt.Sprintf("Error creating channel: %v", err), extra)
		return err
	}

	queueDeclare, err := channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.log.Error(
			logger.Queue,
			logger.RabbitMQRegisterConsumer,
			fmt.Sprintf("Error QueueDeclare channel: %v", err),
			extra,
		)
		return err
	}

	deliveries, err := channel.Consume(
		queueDeclare.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.log.Error(
			logger.Queue,
			logger.RabbitMQRegisterConsumer,
			fmt.Sprintf("Error Consume channel: %v", err),
			extra,
		)
		return err
	}

	go func() {
		for delivery := range deliveries {
			extra = map[logger.ExtraKey]interface{}{
				logger.QueueName: name,
				logger.Body:      string(delivery.Body),
			}
			if err = callback(delivery.Body); err != nil {
				if err = delivery.Nack(false, false); err != nil {
					r.log.Error(
						logger.Queue,
						logger.RabbitMQRegisterConsumer,
						fmt.Sprintf("Error Consume message: %v", err),
						extra,
					)
				}
				r.log.Error(
					logger.Queue,
					logger.RabbitMQRegisterConsumer,
					fmt.Sprintf("Error Consume message: %v", err),
					extra,
				)
			} else {
				if err = delivery.Ack(false); err != nil {
					r.log.Error(
						logger.Queue,
						logger.RabbitMQRegisterConsumer,
						fmt.Sprintf("Error Ack Consume message: %v", err),
						extra,
					)
				}
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) handleReconnect() {
	for {
		err := <-r.notifyClose
		if err != nil {
			r.log.Error(logger.Queue, logger.RabbitMQ, fmt.Sprintf("Connection lost: %v", err), nil)

			for {
				r.log.Info(logger.Queue, logger.RabbitMQ, "Attempting to reconnect to RabbitMQ...", nil)
				time.Sleep(5 * time.Second)

				conn, dialErr := amqp.Dial(r.url)
				if dialErr == nil {
					r.mu.Lock()
					r.conn = conn
					r.notifyClose = conn.NotifyClose(make(chan *amqp.Error))
					r.mu.Unlock()

					r.log.Info(logger.Queue, logger.RabbitMQ, "Successfully reconnected to RabbitMQ", nil)

					r.recoverConsumers()
					break
				}

				r.log.Error(logger.Queue, logger.RabbitMQ, fmt.Sprintf("Reconnect failed: %v", err), nil)
			}
		}
	}
}

func (r *RabbitMQ) recoverConsumers() {
	r.consumerLock.Lock()
	defer r.consumerLock.Unlock()

	for name, callback := range r.consumers {
		r.log.Info(logger.Queue, logger.RabbitMQRegisterConsumer, fmt.Sprintf("Re-registering consumer: %s", name), nil)
		if err := r.setupConsumer(name, callback); err != nil {
			r.log.Error(logger.Queue, logger.RabbitMQRegisterConsumer, fmt.Sprintf("Failed to re-register consumer %s: %v", name, err), nil)
		}
	}
}
