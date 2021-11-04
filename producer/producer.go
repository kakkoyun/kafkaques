package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/twmb/franz-go/pkg/kgo"
)

func Run(ctx context.Context, logger log.Logger, brokers []string, topic string) error {
	logger = log.With(logger, "component", "producer")

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.ClientID("kafkaques"),
		kgo.RecordDeliveryTimeout(2*time.Second),
		kgo.RetryTimeout(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer client.Close()
	defer func() {
		if err := client.Flush(ctx); err != nil {
			level.Error(logger).Log("msg", "failed to flush", "err", err)
		}
	}()

	level.Info(logger).Log("msg", "producer created")
	defer level.Info(logger).Log("msg", "producer exited")

	i := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Print(".")
			val := fmt.Sprintf("Hello! count: %d", i)
			rec := kgo.StringRecord(val)

			// p.Produce(ctx, rec, func(_ *Record, err error) {
			// 	// defer wg.Done()

			// 	if err != nil {
			// 		level.Error(logger).Log(
			// 			"msg", "failed to deliver message",
			// 			"topic", rec.Topic,
			// 			"partition", rec.Partition,
			// 			"headers", rec.Headers,
			// 			"err", err,
			// 		)
			// 	} else {
			// 		level.Info(logger).Log(
			// 			"msg", "message delivered",
			// 			"topic", rec.Topic,
			// 			"partition", rec.Partition,
			// 			"headers", rec.Headers,
			// 		)
			// 	}
			// })

			if err := client.ProduceSync(ctx, rec).FirstErr(); err != nil {
				level.Error(logger).Log(
					"msg", "failed to deliver message",
					"topic", rec.Topic,
					"partition", rec.Partition,
					"offset", rec.Offset,
					"headers", fmt.Sprintf("%+v", rec.Headers),
					"err", err,
				)
			} else {
				level.Info(logger).Log(
					"msg", "message delivered",
					"topic", rec.Topic,
					"partition", rec.Partition,
					"offset", rec.Offset,
					"headers", rec.Headers,
				)
			}
		}
		i++
	}
}
