package producer

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func Run(ctx context.Context, logger log.Logger, brokers, topic string) error {
	logger = log.With(logger, "component", "producer")

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer p.Close()

	level.Info(logger).Log("msg", "producer created", "producer", p)

	errChan := make(chan error)
	go func() {
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case e := <-p.Events():
				switch m := e.(type) {
				case *kafka.Message:
					if m.TopicPartition.Error != nil {
						level.Error(logger).Log(
							"msg", "delivery failed",
							"partition", m.TopicPartition,
							"err", m.TopicPartition,
							"headers", m.Headers,
						)
					} else {
						level.Info(logger).Log(
							"msg", "message delivered",
							"partition", m.TopicPartition,
							"offset", m.TopicPartition.Offset,
							"topic", m.TopicPartition.Topic,
							"headers", m.Headers,
						)
					}
				case kafka.Error:
					level.Error(logger).Log("msg", "failed to send", "code", m.Code(), "err", m)
					if m.Code() == kafka.ErrAllBrokersDown {
						errChan <- fmt.Errorf("failed to continue, code %s: %v", m.Code(), m)
						return
					}
				default:
					level.Warn(logger).Log("msg", "ignored", "message", e)
				}
			}
		}
	}()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		default:
			value := fmt.Sprintf("Hello! count: %d", i)
			p.ProduceChannel() <- &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Value: []byte(value),
			}
		}
		i++
	}
}
