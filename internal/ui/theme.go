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
	colors := map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:         color.NRGBA{R: 17, G: 19, B: 24, A: 255},
		theme.ColorNameForeground:         color.NRGBA{R: 241, G: 242, B: 245, A: 255},
		theme.ColorNamePrimary:            color.NRGBA{R: 184, G: 58, B: 70, A: 255},
		theme.ColorNameButton:             color.NRGBA{R: 34, G: 39, B: 51, A: 255},
		theme.ColorNameDisabledButton:     color.NRGBA{R: 42, G: 46, B: 56, A: 255},
		theme.ColorNameDisabled:           color.NRGBA{R: 105, G: 111, B: 123, A: 255},
		theme.ColorNameHover:              color.NRGBA{R: 206, G: 75, B: 87, A: 55},
		theme.ColorNameFocus:              color.NRGBA{R: 206, G: 75, B: 87, A: 105},
		theme.ColorNameInputBackground:    color.NRGBA{R: 27, G: 31, B: 39, A: 255},
		theme.ColorNameHeaderBackground:   color.NRGBA{R: 27, G: 31, B: 39, A: 255},
		theme.ColorNameMenuBackground:     color.NRGBA{R: 27, G: 31, B: 39, A: 255},
		theme.ColorNameOverlayBackground:  color.NRGBA{R: 27, G: 31, B: 39, A: 255},
		theme.ColorNamePlaceHolder:        color.NRGBA{R: 146, G: 153, B: 167, A: 255},
		theme.ColorNameSeparator:          color.NRGBA{R: 48, G: 54, B: 66, A: 255},
		theme.ColorNameInputBorder:        color.NRGBA{R: 48, G: 54, B: 66, A: 255},
		theme.ColorNamePressed:            color.NRGBA{R: 91, G: 39, B: 47, A: 110},
		theme.ColorNameScrollBar:           color.NRGBA{R: 82, G: 89, B: 103, A: 255},
		theme.ColorNameSelection:           color.NRGBA{R: 91, G: 39, B: 47, A: 255},
		theme.ColorNameShadow:              color.NRGBA{R: 0, G: 0, B: 0, A: 90},
		theme.ColorNameSuccess:             color.NRGBA{R: 85, G: 184, B: 138, A: 255},
		theme.ColorNameWarning:             color.NRGBA{R: 212, G: 173, B: 114, A: 255},
		theme.ColorNameError:               color.NRGBA{R: 224, G: 100, B: 107, A: 255},
		theme.ColorNameForegroundOnPrimary: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
	}
	if value, ok := colors[name]; ok {
		return value
	}
	return t.base.Color(name, theme.VariantDark)
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
