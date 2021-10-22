package producer

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/kakkoyun/kafkaques/kafkaques"
)

func Run(ctx context.Context, logger log.Logger, flags kafkaques.ProducerFlags) error {
	logger = log.With(logger, "component", "producer")

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": flags.Broker})
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer p.Close()

	level.Info(logger).Log("msg", "producer created", "producer", p)

	go func() {
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
		default:
			value := fmt.Sprintf("Hello! count: %d", i)
			p.ProduceChannel() <- &kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &flags.Topic,
					Partition: kafka.PartitionAny},
				Value: []byte(value),
			}
		}
		i++
	}
}
