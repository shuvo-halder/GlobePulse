package rabbitmq

import (
	"github.com/global-news/analytics-service/pkg/logger"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	logger.Log.Info("Connected to RabbitMQ", zap.String("url", url))
	return &Client{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (c *Client) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.Conn != nil {
		c.Conn.Close()
	}
}
