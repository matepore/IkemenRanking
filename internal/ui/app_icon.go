package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed app_icon.svg
var appIconSVG []byte

// AppIcon returns the embedded application icon.
func AppIcon() fyne.Resource {
	return fyne.NewStaticResource("app_icon.svg", appIconSVG)
}
