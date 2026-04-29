package information

import (
	"fyne.io/fyne/v2"
)

const (
	width  = 300
	height = 200
)

type Window struct {
	app    fyne.App
	window fyne.Window

	title string
}

func New(app fyne.App, title string) *Window {
	return &Window{app: app, title: title}
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
