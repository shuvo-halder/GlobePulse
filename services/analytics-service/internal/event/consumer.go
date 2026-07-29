package event

import (
	"context"
	"encoding/json"

	"github.com/global-news/analytics-service/internal/domain"
	"github.com/global-news/analytics-service/pkg/logger"
	"github.com/global-news/analytics-service/pkg/rabbitmq"
	types "github.com/global-news/shared-types"
	"go.uber.org/zap"
)

type Consumer struct {
	client  *rabbitmq.Client
	service domain.AnalyticsUseCase
}

func NewConsumer(client *rabbitmq.Client, service domain.AnalyticsUseCase) *Consumer {
	return &Consumer{
		client:  client,
		service: service,
	}
}

func (c *Consumer) StartConsuming(ctx context.Context, queueName string) error {
	q, err := c.client.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.client.Channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	logger.Log.Info("Started consuming messages", zap.String("queue", queueName))

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("Stopping consumer due to context cancellation")
				return
			case d, ok := <-msgs:
				if !ok {
					logger.Log.Warn("Message channel closed")
					return
				}

				var event types.AnalyticsEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					logger.Log.Error("Failed to unmarshal event", zap.Error(err))
					d.Nack(false, false)
					continue
				}

				if err := c.service.ProcessEvent(ctx, &event); err != nil {
					logger.Log.Error("Failed to process event", zap.Error(err))
					d.Nack(false, true)
				} else {
					d.Ack(false)
				}
			}
		}
	}()

	return nil
}
