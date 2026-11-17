package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewMainWindow builds the initial application window.
func NewMainWindow(application fyne.App) fyne.Window {
	window := application.NewWindow("IkemenRanking")
	window.Resize(fyne.NewSize(800, 520))

	title := widget.NewLabelWithStyle(
		"IkemenRanking",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	description := widget.NewLabel(
		"Automatización de combates CPU vs CPU y generación de rankings para Ikemen GO.",
	)
	status := widget.NewLabel("Estado: estructura inicial de la aplicación.")

	window.SetContent(container.NewVBox(
		title,
		widget.NewSeparator(),
		description,
		status,
	))

	return window
}
