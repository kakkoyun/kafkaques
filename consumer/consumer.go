package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/twmb/franz-go/pkg/kgo"
)

func Run(ctx context.Context, logger log.Logger, brokers []string, group string, topic string) error {
	logger = log.With(logger, "component", "consumer")
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ClientID("kafkaques"),
		kgo.RetryTimeout(5 * time.Second),
		// kgo.DisableAutoCommit(),
	}
	if group != "" {
		opts = append(opts, kgo.ConsumerGroup(group))
	}
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer client.Close()

	level.Info(logger).Log("msg", "consumer created")
	defer level.Info(logger).Log("msg", "consumer exited")

consumerLoop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return errors.New("client closed")
			}

			for _, err := range fetches.Errors() {
				level.Error(logger).Log(
					"msg", "failed to consume",
					"topic", err.Topic,
					"partition", err.Partition,
					"err", err,
				)
				break consumerLoop
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				rec := iter.Next()
				level.Info(logger).Log(
					"msg", "consumed record",
					"topic", rec.Topic,
					"partition", rec.Partition,
					"offset", rec.Offset,
					"headers", fmt.Sprintf("%+v", rec.Headers),
					"message", string(rec.Value),
				)
			}
		}
	}

	return nil
}
