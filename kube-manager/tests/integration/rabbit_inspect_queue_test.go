package integration

import (
	"cmp"
	"context"
	"os"
	"testing"
	"time"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const publishedMessages = 5

func TestRabbitInspectQueue(t *testing.T) {
	addr := os.Getenv("RABBIT_ADDR")
	if addr == "" {
		t.Skip("RABBIT_ADDR is not set")
	}

	user := cmp.Or(os.Getenv("RABBIT_USER"), "guest")
	password := cmp.Or(os.Getenv("RABBIT_PASSWORD"), "guest")
	queue := "kube-manager-test-" + security.RandAlphaNum(10)

	producer, err := rabbit.NewMessageBroker(addr, user, password, queue)
	require.NoError(t, err)

	t.Cleanup(func() {
		func() {
			producer.RLock()
			defer producer.RUnlock()

			if _, err := producer.Channel.QueueDelete(queue, false, false, false); err != nil {
				// A leaked durable queue outlives the run on a shared broker.
				t.Logf("delete queue %q: %v", queue, err)
			}
		}()

		require.NoError(t, producer.Close())
	})

	for range publishedMessages {
		require.NoError(t, producer.Publish(t.Context(), &api.Sort{Field: "name", Key: "asc"}))
	}

	requireQueueState(t, producer, publishedMessages, 0)

	consumer, err := rabbit.NewMessageBroker(addr, user, password, queue, rabbit.WithConsumer("inspect-queue-test", publishedMessages))
	require.NoError(t, err)

	// Everything is prefetched and unacked, so nothing counts towards the backlog.
	requireQueueState(t, producer, 0, 1)

	for range publishedMessages {
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		msg := &api.Sort{}
		require.NoError(t, consumer.Consume(ctx, msg))
		require.Equal(t, "name", msg.GetField())
		cancel()
	}

	requireQueueState(t, producer, 0, 1)

	require.NoError(t, consumer.Close())

	// A count climbing back to 1 would mean Close left the reconnect loop redialing.
	requireQueueState(t, producer, 0, 0)

	// A passive inspect of a missing queue kills the channel; the broker must recover on its own.
	func() {
		producer.RLock()
		defer producer.RUnlock()

		_, err := producer.Channel.QueueDelete(queue, false, false, false)
		require.NoError(t, err)
	}()

	_, _, err = producer.InspectQueue(t.Context())
	require.Error(t, err)

	// The reconnect redeclares the queue the delete took away, with no consumer left on it.
	requireQueueState(t, producer, 0, 0)
}

func requireQueueState(t *testing.T, mb *rabbit.MessageBroker, wantMessages, wantConsumers int) {
	t.Helper()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		messages, consumers, err := mb.InspectQueue(t.Context())
		assert.NoError(c, err)
		assert.Equal(c, wantMessages, messages)
		assert.Equal(c, wantConsumers, consumers)
	}, 30*time.Second, 200*time.Millisecond)
}
