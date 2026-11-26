//go:build windows

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// Use Fyne's built-in dialogs on Windows instead of the third-party COM
// wrapper, whose IFileOpenDialog.Release call crashes in the user's runtime.
func showFolderPicker(parent fyne.Window, _ string, onPicked func(string), onError func(error)) {
	picker := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil { onError(err); return }
		if uri != nil { onPicked(uri.Path()) }
	}, parent)
	picker.Show()
}

func showExecutablePicker(parent fyne.Window, _ string, onPicked func(string), onError func(error)) {
	picker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil { onError(err); return }
		if reader == nil { return }
		path := reader.URI().Path()
		if err := reader.Close(); err != nil { onError(err); return }
		onPicked(path)
	}, parent)
	picker.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
	picker.Show()
}
