package information

import (
	"fyne.io/fyne/v2"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

const (
	width  = 300
	height = 200
)

type Window struct {
	app    fyne.App
	window fyne.Window

	title    string
	useCases interfaces.UseCases
}

func New(app fyne.App, title string, useCases interfaces.UseCases) *Window {
	return &Window{app: app, useCases: useCases, title: title}
}

func (w *Window) Build(content fyne.CanvasObject) {
	window := w.app.NewWindow(w.title)
	window.Resize(fyne.NewSize(width, height))
	window.SetContent(content)

	w.window = window
}

func (w *Window) Show() {
	if w.window == nil {
		return
	}

	w.window.Show()
}

func (w *Window) Close() {
	if w.window == nil {
		return
	}

	w.window.Close()
	w.window = nil
}
