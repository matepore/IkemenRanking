package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

// NewMainWindow builds the main application window.
func NewMainWindow(application fyne.App) fyne.Window {
	window := application.NewWindow("IkemenRanking")
	window.Resize(fyne.NewSize(900, 620))

	title := widget.NewLabelWithStyle(
		"IkemenRanking",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	description := widget.NewLabel(
		"Automatización de combates CPU vs CPU y generación de rankings para Ikemen GO.",
	)
	installationPath := widget.NewLabel("Instalación: ninguna carpeta seleccionada.")
	characterCount := widget.NewLabel("Personajes encontrados: 0")
	status := widget.NewLabel("Seleccioná la carpeta raíz de tu instalación de Ikemen GO.")
	characters := make([]string, 0)

	characterList := widget.NewList(
		func() int { return len(characters) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(item widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(characters[item])
		},
	)

	loadInstallation := func(path string) {
		loaded, err := ikemen.LoadCharacters(path)
		if err != nil {
			status.SetText("No se pudo cargar la instalación.")
			dialog.ShowError(err, window)
			return
		}

		characters = loaded
		installationPath.SetText(fmt.Sprintf("Instalación: %s", path))
		characterCount.SetText(fmt.Sprintf("Personajes encontrados: %d", len(characters)))
		if len(characters) == 0 {
			status.SetText("Se leyó data/select.def, pero no hay personajes en [Characters].")
		} else {
			status.SetText("Lista de personajes cargada correctamente.")
		}
		characterList.Refresh()
	}

	selectFolder := widget.NewButton("Seleccionar carpeta de Ikemen GO", func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri == nil {
				return
			}
			loadInstallation(uri.Path())
		}, window)
		folderDialog.Show()
	})

	header := container.NewVBox(
		title,
		widget.NewSeparator(),
		description,
		selectFolder,
		installationPath,
		characterCount,
		status,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Identificadores de personajes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	window.SetContent(container.NewBorder(header, nil, nil, nil, characterList))

	return window
}
