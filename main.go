package main

import (
	"github.com/presskey/rabbit/internal/app"
	"github.com/presskey/rabbit/internal/queue"
)

func main() {
	queueConnection, _ := queue.GetConn("amqp://localhost:5672/")
	defer queueConnection.Connection.Close()

	app := app.NewRabbit(queueConnection)
	app.Start()
}
