package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

const RABBITMQ_LOCAL_URL = "amqp://localhost:5672/"

type Rabbit struct {
	QueueConnection queue.Conn
	App             fyne.App
	Log             []string
	LogsWindow      fyne.Window
}

func NewRabbit() *Rabbit {
	app := app.NewWithID("com.example.rabbit")

	return &Rabbit{
		QueueConnection: queue.Conn{},
		App:             app,
		Log:             []string{},
		LogsWindow:      nil,
	}
}

func (r *Rabbit) Start() {
	r.App.Settings().SetTheme(theme.DarkTheme())

	mainWindow := r.App.NewWindow("🐇")
	mainWindow.SetMaster()
	mainWindow.SetFixedSize(true)
	mainWindow.CenterOnScreen()
	mainWindow.Resize(fyne.NewSize(400, 530))

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
			errorDialog := dialog.NewError(errors.New("can't connect to RabbitMQ"), mainWindow)
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
			errorDialog := dialog.NewError(err, mainWindow)
			errorDialog.Show()
		} else {
			r.AddLog(fmt.Sprintf("-> %s (%s)", exchange, key))
		}
	})

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ListIcon(), func() {
			for _, w := range r.App.Driver().AllWindows() {
				if w.Title() == "[logs]" {
					return
				}
			}

			r.LogsWindow = r.App.NewWindow("[logs]")
			r.LogsWindow.Resize(fyne.NewSize(400, 400))

			list := widget.NewList(
				func() int {
					return len(r.Log)
				},
				func() fyne.CanvasObject {
					return widget.NewLabel("template")
				},
				func(i widget.ListItemID, o fyne.CanvasObject) {
					o.(*widget.Label).SetText(r.Log[i])
				},
			)

			r.LogsWindow.SetContent(list)
			r.LogsWindow.Show()
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			for _, w := range r.App.Driver().AllWindows() {
				if w.Title() == "[settings]" {
					return
				}
			}

			settingsWindow := r.App.NewWindow("[settings]")
			settingsWindow.Resize(fyne.NewSize(400, 85))

			urlBinding := binding.NewString()
			urlLabel := widget.NewLabel("RabbitMQ URL")
			urlEntry := widget.NewEntryWithData(urlBinding)
			url := r.App.Preferences().StringWithFallback("RabbitMQUrl", RABBITMQ_LOCAL_URL)
			urlBinding.Set(url)

			saveButton := widget.NewButton("save", func() {
				url, _ := urlBinding.Get()
				if url == "" {
					url = RABBITMQ_LOCAL_URL
				}
				r.App.Preferences().SetString("RabbitMQUrl", url)
				r.QueueConnection, _ = queue.GetConn(url)
				settingsWindow.Close()
			})

			settingsContainer := container.New(
				layout.NewVBoxLayout(),
				container.New(layout.NewFormLayout(), urlLabel, urlEntry),
				saveButton,
			)

			settingsWindow.SetContent(settingsContainer)
			settingsWindow.Show()
		}),
	)

	content := container.New(
		layout.NewVBoxLayout(),
		toolbar,
		input,
		container.NewGridWithColumns(2, widget.NewLabel("exchange"), exchangeEntry, widget.NewLabel("routing key"), keyEntry),
		sendButton,
	)

	mainWindow.SetContent(content)

	conn, err := queue.GetConn(r.App.Preferences().StringWithFallback("RabbitMQUrl", RABBITMQ_LOCAL_URL))
	if err != nil {
		errorDialog := dialog.NewError(err, mainWindow)
		errorDialog.Show()
	} else {
		r.AddLog("connected to RabbitMQ")
		r.QueueConnection = conn
		defer r.QueueConnection.Connection.Close()
	}

	mainWindow.ShowAndRun()
}

func (r *Rabbit) AddLog(s string) {
	r.Log = append(r.Log, fmt.Sprintf("[%v] %s", time.Now().Format(time.TimeOnly), s))
	if r.LogsWindow != nil {
		r.LogsWindow.Content().Refresh()
	}
}
