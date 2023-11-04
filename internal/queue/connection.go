package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Conn struct
type Conn struct {
	Channel    *amqp.Channel
	Connection *amqp.Connection
}

func GetConn(rabbitURL string) (Conn, error) {
	conn, err := amqp.Dial(rabbitURL)

	if err != nil {
		return Conn{}, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return Conn{}, err
	}

	return Conn{
		Channel:    ch,
		Connection: conn,
	}, nil
}
