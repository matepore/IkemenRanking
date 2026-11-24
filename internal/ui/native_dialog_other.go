//go:build !windows

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func showFolderPicker(parent fyne.Window, _ string, onPicked func(string), onError func(error)) {
	d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil { onError(err); return }
		if uri != nil { onPicked(uri.Path()) }
	}, parent)
	d.Show()
}

func showExecutablePicker(parent fyne.Window, _ string, onPicked func(string), onError func(error)) {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil { onError(err); return }
		if reader == nil { return }
		path := reader.URI().Path()
		if err := reader.Close(); err != nil { onError(err); return }
		onPicked(path)
	}, parent)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
	d.Show()
}
