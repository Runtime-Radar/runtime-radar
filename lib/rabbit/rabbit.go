package rabbit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"google.golang.org/protobuf/proto"
)

const (
	connectTimeout = time.Second * 5
	reconnectDelay = time.Second * 3
)

var (
	ErrClosedUnexpectedly = errors.New("channel closed unexpectedly")
	ErrNotInitialized     = errors.New("not initialized")
	ErrReconnecting       = errors.New("broker is reconnecting")
)

type PublishConsumer interface {
	Publish(ctx context.Context, in proto.Message) error
	Consume(ctx context.Context, out proto.Message) error
}

type MessageBroker struct {
	sync.RWMutex

	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue

	Msgs <-chan amqp.Delivery

	withConsumer  bool
	consumerName  string
	prefetchCount int

	closed atomic.Bool

	stateReporter func(bool)
}

type ConfigOption func(*MessageBroker)

func WithConsumer(name string, prefetchCount int) ConfigOption {
	return func(mb *MessageBroker) {
		mb.withConsumer = true
		// Consumer name must be unique
		mb.consumerName = name + "-" + security.RandAlphaNum(5)
		mb.prefetchCount = prefetchCount
	}
}

func WithStateReporter(f func(bool)) ConfigOption {
	return func(mb *MessageBroker) {
		mb.stateReporter = f
	}
}

func NewMessageBroker(address, user, password, queue string, opts ...ConfigOption) (*MessageBroker, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s", user, url.QueryEscape(password), address)

	mb := &MessageBroker{}

	for _, opt := range opts {
		opt(mb)
	}

	if err := mb.init(url, queue); err != nil {
		return nil, err
	}

	go func() {
		for {
			notifyClose := make(chan *amqp.Error)
			mb.Connection.NotifyClose(notifyClose)
			mb.reportState(true)

			nc, ok := <-notifyClose
			mb.Lock()
			mb.reportState(false)

			if mb.closed.Load() {
				mb.Unlock()

				return
			}

			// According to documentation channel will be closed only on clean shutdown
			reason := "shutdown"
			if ok {
				reason = nc.Reason
			}

			t0 := time.Now()
			for {
				if err := mb.init(url, queue); err != nil {
					log.Error().Err(err).Str("reason", reason).Msgf("RabbitMQ connection reopen failure")
					time.Sleep(reconnectDelay)
					continue
				}

				break
			}

			mb.Unlock()
			log.Info().Str("reason", reason).Str("delay", time.Since(t0).String()).Msgf("RabbitMQ connection reopened")
		}
	}()

	return mb, nil
}

func (mb *MessageBroker) init(url, queue string) error {
	conn, err := amqp.DialConfig(url, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, connectTimeout)
		},
	})
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	mb.Connection = conn
	mb.Channel = ch
	mb.Queue = &q

	if mb.withConsumer {
		if err := mb.Channel.Qos(mb.prefetchCount, 0, false); err != nil {
			return err
		}

		mb.Msgs, err = mb.Channel.Consume(queue, mb.consumerName, false, false, false, false, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close does not wait out a reconnect: it returns ErrReconnecting with the broker still open, so
// SIGTERM can't hang on a broker outage.
func (mb *MessageBroker) Close() error {
	if !mb.TryRLock() {
		return ErrReconnecting
	}
	defer mb.RUnlock()

	// Set before the close fires NotifyClose, or the reconnect loop redials and redeclares the queue.
	mb.closed.Store(true)

	if mb.Connection == nil {
		return ErrNotInitialized
	}

	if err := mb.Connection.Close(); err != nil {
		return err
	}

	return mb.Channel.Close()
}

func (mb *MessageBroker) Publish(ctx context.Context, in proto.Message) error {
	mb.RLock()
	defer mb.RUnlock()

	b, err := proto.Marshal(in)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/octet-stream", // alternative is "application/protobuf", according to expired IETF draft
		Body:         b,
	}

	return mb.Channel.PublishWithContext(ctx, "", mb.Queue.Name, false, false, msg)
}

func (mb *MessageBroker) Consume(ctx context.Context, out proto.Message) error {
	mb.RLock()
	defer mb.RUnlock()

	if mb.Msgs == nil {
		return ErrNotInitialized
	}

	select {
	case msg, ok := <-mb.Msgs:
		if !ok {
			return ErrClosedUnexpectedly
		}

		if err := proto.Unmarshal(msg.Body, out); err != nil {
			return err
		}

		return msg.Ack(false)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// InspectQueue passively inspects the queue and reports the number of messages ready for delivery
// and the number of attached consumers. Like Close, it returns ErrReconnecting rather than blocking
// while the reconnect loop holds the lock.
func (mb *MessageBroker) InspectQueue(ctx context.Context) (messages, consumers int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}

	if !mb.TryRLock() {
		return 0, 0, ErrReconnecting
	}
	defer mb.RUnlock()

	if mb.Channel == nil || mb.Queue == nil {
		return 0, 0, ErrNotInitialized
	}

	q, err := mb.Channel.QueueDeclarePassive(mb.Queue.Name, true, false, false, false, nil)
	if err != nil {
		// A failed passive declare kills the channel, which only the reconnect loop recreates.
		_ = mb.Connection.Close()

		return 0, 0, fmt.Errorf("inspect queue %q: %w", mb.Queue.Name, err)
	}

	return q.Messages, q.Consumers, nil
}

func (mb *MessageBroker) reportState(isAlive bool) {
	if mb.stateReporter != nil {
		mb.stateReporter(isAlive)
	}
}
