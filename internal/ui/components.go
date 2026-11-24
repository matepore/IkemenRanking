package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// newSectionCard groups a related part of the workflow in a consistent panel.
func newSectionCard(title, subtitle string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, subtitle, content)
}

// pathEntry validates its contents after submission or when it loses focus.
type pathEntry struct {
	*widget.Entry
	onFocusLost func()
}

func newPathEntry(placeholder string) *pathEntry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return &pathEntry{Entry: entry}
}

func (e *pathEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.onFocusLost != nil {
		fyne.Do(e.onFocusLost)
	}
}
