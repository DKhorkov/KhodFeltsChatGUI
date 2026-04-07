package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type CustomEntry struct {
	widget.Entry

	OnSubmit func(string)
}

func NewCustomEntry() *CustomEntry {
	e := &CustomEntry{}
	e.ExtendBaseWidget(e)

	return e
}

func (e *CustomEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if d, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
			if d.CurrentKeyModifiers()&fyne.KeyModifierShift != 0 {
				e.TypedRune('\n')

				return
			}
		}

		if e.OnSubmit != nil {
			e.OnSubmit(e.Text)
		}

		return
	}

	e.Entry.TypedKey(key)
}
