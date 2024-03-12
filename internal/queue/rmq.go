package queue

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tel4vn/fins-microservices/common/log"
	"github.com/tel4vn/fins-microservices/common/util"
)

type Rcfg struct {
	Address  string
	Username string
	Password string
	DB       int
}

var RMQ *RMQConnection

var Client *RMQClient

const (
	tag           = "rmq"
	pollDuration  = 100 * time.Millisecond
	prefetchLimit = 1000
)

type RMQConnection struct {
	RedisClient *redis.Client
	Config      Rcfg
	Conn        rmq.Connection
	Queues      map[string]rmq.Queue
	Server      *RMQServer
}

type RMQServer struct {
	mu     sync.Mutex
	conn   rmq.Connection
	Queues map[string]rmq.Queue
}

type RMQClient struct {
	conn rmq.Connection
}

type ConsumerConfig struct {
	Name           string
	Handler        rmq.ConsumerFunc
	NumberInstance int
}

func NewConsumerConfig(name string, handler rmq.ConsumerFunc, numberInstance int) ConsumerConfig {
	return ConsumerConfig{Name: name, Handler: handler, NumberInstance: numberInstance}
}

func NewRMQ(config Rcfg) *RMQConnection {
	poolSize := runtime.NumCPU() * 4
	errChan := make(chan error, 10)
	go logErrors(errChan)
	client := redis.NewClient(&redis.Options{
		Addr:            config.Address,
		Password:        config.Password,
		DB:              config.DB,
		PoolSize:        poolSize,
		PoolTimeout:     time.Duration(20) * time.Second,
		ReadTimeout:     time.Duration(20) * time.Second,
		WriteTimeout:    time.Duration(20) * time.Second,
		ConnMaxIdleTime: time.Duration(20) * time.Second,
	})
	connection, err := rmq.OpenConnectionWithRedisClient(tag, client, errChan)
	if err != nil {
		log.Fatal(err)
	}
	return &RMQConnection{
		RedisClient: client,
		Config:      config,
		Conn:        connection,
	}
}
func logErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Error("heartbeat error (limit): ", err)
			} else {
				log.Error("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Error("consume error: ", err)
		case *rmq.DeliveryError:
			log.Error("delivery error: ", err.Delivery, err)
		default:
			log.Error("other error: ", err)
		}
	}
}

func (c *RMQConnection) NewServer() error {
	c.Server = &RMQServer{
		conn:   c.Conn,
		Queues: make(map[string]rmq.Queue),
	}
	return nil
}

func (c *RMQConnection) NewClient() (*RMQClient, error) {
	return &RMQClient{
		conn: c.Conn,
	}, nil
}

var (
	ERR_QUEUE_IS_EXIST     = errors.New("queue is existed")
	ERR_QUEUE_IS_NOT_EXIST = errors.New("queue is not existed")
)

func (srv *RMQServer) AddQueue(name string, handler rmq.ConsumerFunc, numConsumers int) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, ok := srv.Queues[name]; ok {
		return ERR_QUEUE_IS_EXIST
	}
	queue, err := srv.conn.OpenQueue(name)
	if err != nil {
		return err
	}
	srv.Queues[name] = queue
	if err := queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		return err
	}
	for i := 0; i < numConsumers; i++ {
		if _, err := queue.AddConsumerFunc(tag, handler); err != nil {
			return err
		}
	}
	return nil
}

type Consumer struct {
}

func (c *Consumer) Consume(delivery rmq.Delivery) {
	log.Println("Received message: ", delivery.Payload())
	err := delivery.Ack()
	if err != nil {
		return
	}
}

func (conn *RMQConnection) Close() {
	<-conn.Conn.StopAllConsuming()
	cleaner := rmq.NewCleaner(conn.Conn)
	i, err := cleaner.Clean()
	if err != nil {
		log.Println(err)
	}
	log.Println("Cleaned", i, "messages")
}

func (srv *RMQServer) Run() {
	srv.waitForSignals()
}

func (srv *RMQServer) waitForSignals() {
	log.Println("Waiting for signals...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	<-srv.conn.StopAllConsuming()
}

func (srv *RMQServer) RemoveQueue(name string) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, ok := srv.Queues[name]; !ok {
		return ERR_QUEUE_IS_NOT_EXIST
	}
	q, err := srv.conn.OpenQueue(name)
	if err != nil {
		return err
	}
	_, _, err = q.Destroy()
	if err != nil {
		return err
	}
	srv.Queues[name].StopConsuming()
	// srv.Queues[name].PurgeReady()
	// srv.Queues[name].PurgeRejected()
	delete(srv.Queues, name)
	return nil
}

func (srv *RMQServer) IsHasQueue(name string) bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	_, ok := srv.Queues[name]
	return ok
}

func (c *RMQClient) Publish(queueName string, payload []byte) error {
	queue, err := c.conn.OpenQueue(queueName)
	if err != nil {
		return err
	}
	return queue.PublishBytes(payload)
}

func UnmarshallMessage[M interface{}](delivery rmq.Delivery) (result M, err error) {
	if err = json.Unmarshal([]byte(delivery.Payload()), &result); err != nil {
		// handle json error
		log.Error(err)
		if err = delivery.Reject(); err != nil {
			// handle reject error
			log.Error(err)
		}
	}
	return
}

func Publish[M interface{}](message M, topic string) {
	messageBytes, err := util.ConvertToBytes(message)
	if err != nil {
		log.Error(err)
	} else {
		err = Client.Publish(topic, messageBytes)
		if err != nil {
			log.Error(err)
		}
	}
}

func BatchInitQueues(configs ...ConsumerConfig) {
	for _, config := range configs {
		err := RMQ.Server.AddQueue(config.Name, config.Handler, config.NumberInstance)
		if err != nil {
			log.Error(err)
		} else {
			log.Debugf("Add queue '%s' successfully", config.Name)
		}
	}
}
