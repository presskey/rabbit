package queue

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (conn Conn) Publish(exchange string, key string, msg amqp.Publishing) error {
	return conn.Channel.PublishWithContext(context.Background(),
		exchange,
		key,
		false,
		false,
		msg)
}
