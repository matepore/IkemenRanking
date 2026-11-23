package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/matepore/IkemenRanking/internal/ui"
)

func main() {
	application := app.NewWithID("com.matepore.ikemenranking")
	window := ui.NewMainWindow(application)
	window.ShowAndRun()
}
