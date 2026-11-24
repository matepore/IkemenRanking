//go:build windows

package ui

import (
	"errors"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/harry1453/go-common-file-dialog/cfd"
	"github.com/harry1453/go-common-file-dialog/cfdutil"
	"golang.org/x/sys/windows"
)

func showFolderPicker(_ fyne.Window, currentPath string, onPicked func(string), onError func(error)) {
	path, err := cfdutil.ShowPickFolderDialog(cfd.DialogConfig{
		Title:               "Select Ikemen GO folder",
		DefaultFolder:       existingDialogDirectory(currentPath),
		ParentWindowHandle: uintptr(windows.GetForegroundWindow()),
	})
	if err != nil {
		if !errors.Is(err, cfd.ErrorCancelled) {
			onError(err)
		}
		return
	}
	if path != "" {
		onPicked(path)
	}
}

func showExecutablePicker(_ fyne.Window, currentPath string, onPicked func(string), onError func(error)) {
	path, err := cfdutil.ShowOpenFileDialog(cfd.DialogConfig{
		Title:               "Select Ikemen GO executable",
		DefaultFolder:       existingDialogDirectory(currentPath),
		ParentWindowHandle: uintptr(windows.GetForegroundWindow()),
		FileFilters: []cfd.FileFilter{
			{
				DisplayName: "Windows executable (*.exe)",
				Pattern:     "*.exe",
			},
		},
	})
	if err != nil {
		if !errors.Is(err, cfd.ErrorCancelled) {
			onError(err)
		}
		return
	}
	if path != "" {
		onPicked(path)
	}
}

func existingDialogDirectory(path string) string {
	if path == "" {
		return ""
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return path
	}
	directory := filepath.Dir(path)
	if info, err := os.Stat(directory); err == nil && info.IsDir() {
		return directory
	}
	return ""
}
