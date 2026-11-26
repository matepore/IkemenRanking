package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed app_icon.png
var appIconPNG []byte

// AppIcon returns the embedded PNG application icon.
func AppIcon() fyne.Resource {
	return fyne.NewStaticResource("app_icon.png", appIconPNG)
}
