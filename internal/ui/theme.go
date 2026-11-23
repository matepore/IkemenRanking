package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// graphiteTheme provides the application's Graphite & Crimson visual palette.
type graphiteTheme struct {
	base fyne.Theme
}

// NewGraphiteTheme returns the dark Graphite & Crimson application theme.
func NewGraphiteTheme() fyne.Theme {
	return graphiteTheme{base: theme.DefaultTheme()}
}

func (t graphiteTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 17, G: 19, B: 24, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 241, G: 242, B: 245, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 184, G: 58, B: 70, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 34, G: 39, B: 51, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 42, G: 46, B: 56, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 105, G: 111, B: 123, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 206, G: 75, B: 87, A: 55}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 206, G: 75, B: 87, A: 105}
	case theme.ColorNameInputBackground, theme.ColorNameHeaderBackground,
		theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 27, G: 31, B: 39, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 146, G: 153, B: 167, A: 255}
	case theme.ColorNameSeparator, theme.ColorNameInputBorder:
		return color.NRGBA{R: 48, G: 54, B: 66, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 91, G: 39, B: 47, A: 110}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 82, G: 89, B: 103, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 91, G: 39, B: 47, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 90}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 85, G: 184, B: 138, A: 255}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 212, G: 173, B: 114, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 224, G: 100, B: 107, A: 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t graphiteTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t graphiteTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t graphiteTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
