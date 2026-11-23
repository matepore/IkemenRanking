package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/matepore/IkemenRanking/internal/ui"
)

func main() {
	application := app.NewWithID("com.matepore.ikemenranking")
	application.Settings().SetTheme(ui.NewGraphiteTheme())
	window := ui.NewMainWindow(application)
	window.ShowAndRun()
}
