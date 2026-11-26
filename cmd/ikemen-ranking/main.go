package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/matepore/IkemenRanking/internal/ui"
)

func main() {
	application := app.NewWithID("com.matepore.ikemenranking")
	application.SetIcon(ui.AppIcon())
	application.Settings().SetTheme(ui.NewGraphiteTheme())
	window := ui.NewMainWindow(application)
	window.SetIcon(ui.AppIcon())
	window.ShowAndRun()
}
