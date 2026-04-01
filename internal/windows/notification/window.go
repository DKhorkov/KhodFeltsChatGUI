package notification

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

const (
	title = "Новое сообщение"

	width  = 300
	height = 200
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases interfaces.UseCases
}

func New(app fyne.App, useCases interfaces.UseCases) *Window {
	return &Window{app: app, useCases: useCases}
}

func (w *Window) Build(message fyne.CanvasObject) {
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))
	window.SetContent(
		container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			message,
		),
	)
	window.CenterOnScreen()

	w.window = window
}

func (w *Window) Show() {
	if w.window == nil {
		return
	}

	w.window.Show()

	// Замыкание на случай, когда пришлют много сообщений и w.window будет менять адрес объекта
	window := w.window

	time.AfterFunc(3*time.Second, func() {
		fyne.Do(func() {
			if window != nil {
				window.Close()
			}
		})
	})
}

func (w *Window) Close() {
	if w.window == nil {
		return
	}

	w.window.Close()
	w.window = nil
}
