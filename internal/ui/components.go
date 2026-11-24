package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// newSectionCard groups a related part of the workflow in a consistent panel.
func newSectionCard(title, subtitle string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, subtitle, content)
}

// pathEntry validates its contents after submission or when it loses focus.
type pathEntry struct {
	*widget.Entry
	onFocusLost func()
}

func newPathEntry(placeholder string) *pathEntry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return &pathEntry{Entry: entry}
}

func (e *pathEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.onFocusLost != nil {
		fyne.Do(e.onFocusLost)
	}
}

// pathRowLayout keeps the Browse button directly beside the editable path.
type pathRowLayout struct {
	gap float32
}

func (l pathRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	entry, browse := objects[0], objects[1]
	browseSize := browse.MinSize()
	entryWidth := size.Width - browseSize.Width - l.gap
	// Cap the field width so Browse stays adjacent rather than being
	// positioned far from the text field on wide windows.
	const maxEntryWidth = float32(560)
	if entryWidth > maxEntryWidth {
		entryWidth = maxEntryWidth
	}
	if entryWidth < 0 {
		entryWidth = 0
	}
	entry.Resize(fyne.NewSize(entryWidth, size.Height))
	browse.Move(fyne.NewPos(entryWidth+l.gap, 0))
	browse.Resize(fyne.NewSize(browseSize.Width, size.Height))
}

func (l pathRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}
	entrySize, browseSize := objects[0].MinSize(), objects[1].MinSize()
	height := entrySize.Height
	if browseSize.Height > height {
		height = browseSize.Height
	}
	return fyne.NewSize(entrySize.Width+l.gap+browseSize.Width, height)
}

// newPathRow creates a flexible row with the path field and its Browse button.
func newPathRow(entry, browse fyne.CanvasObject) *fyne.Container {
	return container.New(pathRowLayout{gap: 8}, entry, browse)
}
