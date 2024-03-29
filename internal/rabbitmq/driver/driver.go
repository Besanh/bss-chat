package rabbitmq

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type Config struct {
	Uri                  string
	ChannelNotifyTimeout time.Duration
	Reconnect            struct {
		Interval   time.Duration
		MaxAttempt int
	}
	IndexName    string
	ExchangeName string
	ExchangeType string
	RoutingKey   string
	QueueName    string
}

type RabbitMQ struct {
	mux                  sync.RWMutex
	config               Config
	dialConfig           amqp.Config
	connection           *amqp.Connection
	ChannelNotifyTimeout time.Duration
	ExchangeName         string
	ExchangeType         string
	RoutingKey           string
	QueueName            string
}

var RabbitConnector *RabbitMQ

func New(config Config) *RabbitMQ {
	return &RabbitMQ{
		config:               config,
		dialConfig:           amqp.Config{},
		ChannelNotifyTimeout: config.ChannelNotifyTimeout,
	}
}

// Connect creates a new connection. Use once at application
// startup.
func (r *RabbitMQ) Connect() error {
	con, err := amqp.DialConfig(r.config.Uri, r.dialConfig)
	if err != nil {
		return err
	}
	r.connection = con
	go r.reconnect()
	return nil
}

// Channel returns a new `*amqp.Channel` instance. You must
// call `defer channel.Close()` as soon as you obtain one.
// Sometimes the connection might be closed unintentionally so
// as a graceful handling, try to connect only once.
func (r *RabbitMQ) Channel() (*amqp.Channel, error) {
	if r.connection == nil {
		if err := r.Connect(); err != nil {
			return nil, errors.New("connection is not open")
		}
	}

	channel, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}

	return channel, nil
}

// Connection exposes the essentials of the current connection.
// You should not normally use this but it is there for special
// use cases.
func (r *RabbitMQ) Connection() *amqp.Connection {
	return r.connection
}

// Shutdown triggers a normal shutdown. Use this when you wish
// to shutdown your current connection or if you are shutting
// down the application.
func (r *RabbitMQ) Shutdown() error {
	if r.connection != nil {
		return r.connection.Close()
	}

	return nil
}

// reconnect reconnects to server if the connection or a channel
// is closed unexpectedly. Normal shutdown is ignored. It tries
// maximum of 7200 times and sleeps half a second in between
// each try which equals to 1 hour.
func (r *RabbitMQ) reconnect() {
WATCH:

	conErr := <-r.connection.NotifyClose(make(chan *amqp.Error))
	if conErr != nil {
		log.Error("CRITICAL: Connection dropped, reconnecting")

		var err error

		for i := 1; i <= r.config.Reconnect.MaxAttempt; i++ {
			r.mux.RLock()
			r.connection, err = amqp.DialConfig(r.config.Uri, r.dialConfig)
			r.mux.RUnlock()

			if err == nil {
				log.Info("INFO: Reconnected")

				goto WATCH
			}

			time.Sleep(r.config.Reconnect.Interval)
		}

		log.Error(errors.Wrap(err, "CRITICAL: Failed to reconnect"))
	} else {
		log.Info("INFO: Connection dropped normally, will not reconnect")
	}
}

func (r *RabbitMQ) Ping() error {
	channel, err := r.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open channel")
	}
	defer channel.Close()
	return nil
}

func (r *RabbitMQ) Setup() error {
	channel, err := r.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open channel")
	}
	defer channel.Close()

	if err := r.declareCreate(channel); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareCreate(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		r.config.ExchangeName,
		r.config.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "failed to declare exchange")
	}

	if _, err := channel.QueueDeclare(
		r.config.QueueName,
		true,
		false,
		false,
		false,
		amqp.Table{"x-queue-mode": "lazy"},
	); err != nil {
		return errors.Wrap(err, "failed to declare queue")
	}

	if err := channel.QueueBind(
		r.config.QueueName,
		r.config.RoutingKey,
		r.config.ExchangeName,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "failed to bind queue")
	}

	return nil
}

func (r *RabbitMQ) Publish(body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Error("RabbitMqRepo - Publish : error : ", err)
		return errors.Wrap(err, "failed to marshal json")
	}
	channel, err := r.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open channel")
	}
	defer channel.Close()
	if err := channel.Confirm(false); err != nil {
		return errors.Wrap(err, "failed to put channel in confirmation mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := channel.PublishWithContext(ctx,
		RabbitConnector.ExchangeName,
		RabbitConnector.RoutingKey,
		true,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(string(jsonBody)),
		},
	); err != nil {
		return errors.Wrap(err, "failed to publish message")
	}
	select {
	case ntf := <-channel.NotifyPublish(make(chan amqp.Confirmation, 1)):
		if !ntf.Ack {
			return errors.New("failed to deliver message to exchange/queue")
		}
	case <-channel.NotifyReturn(make(chan amqp.Return)):
		return errors.New("failed to deliver message to exchange/queue")
	case <-time.After(r.ChannelNotifyTimeout):
		log.Println("message delivery confirmation to exchange/queue timed out")
	}

	return nil
}
