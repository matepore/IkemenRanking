package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// newSectionCard groups a related part of the workflow in a consistent panel.
func newSectionCard(title, subtitle string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, subtitle, content)
}
