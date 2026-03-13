package interfaces

import "fyne.io/fyne/v2"

//go:generate mockgen -source=window.go -destination=../../mocks/window/window.go -package=mockwindow -exclude_interfaces=
type Window interface {
	Build(content fyne.CanvasObject)
	Show()
	Close()
}
