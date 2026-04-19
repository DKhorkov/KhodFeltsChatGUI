package entries

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type MultilineEntry struct {
	widget.Entry
}

func NewMultiLineEntry() *MultilineEntry {
	e := &MultilineEntry{}
	e.MultiLine = true

	e.ExtendBaseWidget(e)

	return e
}

func (e *MultilineEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if d, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
			if d.CurrentKeyModifiers()&fyne.KeyModifierShift != 0 {
				e.TypedRune('\n')

				return
			}
		}

		if e.OnSubmitted != nil {
			e.OnSubmitted(e.Text)
		}

		return
	}

	e.Entry.TypedKey(key)
}
