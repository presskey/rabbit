package app

import (
	"encoding/json"
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/presskey/rabbit/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	QueueConnection queue.Conn
}

func NewRabbit(queueConnection queue.Conn) *Rabbit {
	return &Rabbit{
		QueueConnection: queueConnection,
	}
}

func (r Rabbit) Start() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	myWindow := myApp.NewWindow("🐇")
	myWindow.SetMaster()
	myWindow.SetFixedSize(true)
	myWindow.CenterOnScreen()
	myWindow.Resize(fyne.NewSize(400, 400))

	messageBinding := binding.NewString()

	input := widget.NewMultiLineEntry()
	input.Bind(messageBinding)
	input.SetMinRowsVisible(18)
	input.Validator = func(s string) error {
		if json.Valid([]byte(s)) {
			return nil
		}

		return errors.New("invalid json")
	}

	exchangeBinding := binding.NewString()
	exchangeEntry := widget.NewEntryWithData(exchangeBinding)
	keyBinding := binding.NewString()
	keyEntry := widget.NewEntryWithData(keyBinding)

	sendButton := widget.NewButton("send", func() {
		if (r.QueueConnection == queue.Conn{}) {
			errorDialog := dialog.NewError(errors.New("can't connect to RabbitMQ"), myWindow)
			errorDialog.Show()
			return
		}

		exchange, _ := exchangeBinding.Get()
		key, _ := keyBinding.Get()
		message, _ := messageBinding.Get()

		err := r.QueueConnection.Publish(exchange, key, amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
		})

		if err != nil {
			errorDialog := dialog.NewError(err, myWindow)
			errorDialog.Show()
		}
	})

	content := container.New(
		layout.NewVBoxLayout(),
		input,
		container.NewGridWithColumns(2, widget.NewLabel("exchange"), exchangeEntry, widget.NewLabel("routing key"), keyEntry),
		sendButton,
	)

	myWindow.SetContent(content)

	myWindow.ShowAndRun()
}
