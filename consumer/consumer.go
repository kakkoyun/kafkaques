package consumer

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func Run(ctx context.Context, logger log.Logger, brokers string, group string, topics ...string) error {
	logger = log.With(logger, "component", "consumer")
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"broker.address.family": "v4",
		"group.id":              group,
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest",
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer c.Close()

	level.Info(logger).Log("msg", "consumer created", "consumer", c)

	if err := c.SubscribeTopics(topics, nil); err != nil {
		return err
	}

outer:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				level.Info(logger).Log(
					"msg", "message received",
					"partition", e.TopicPartition,
					"message", string(e.Value),
					"headers", e.Headers,
				)
			case kafka.Error:
				level.Error(logger).Log("msg", "failed to receive", "code", e.Code(), "err", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					break outer
				}
			default:
				level.Warn(logger).Log("msg", "ignored", "message", e)
			}
		}
	}

	return nil
}
