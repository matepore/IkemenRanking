//go:build windows

package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"github.com/go-ole/go-ole"
	"github.com/harry1453/go-common-file-dialog/cfd"
	"github.com/harry1453/go-common-file-dialog/cfdutil"
	"golang.org/x/sys/windows"
)

func showFolderPicker(_ fyne.Window, currentPath string, onPicked func(string), onError func(error)) {
	runNativeDialog(func() (string, error) {
		return cfdutil.ShowPickFolderDialog(cfd.DialogConfig{
			Title:               "Select Ikemen GO folder",
			DefaultFolder:       existingDialogDirectory(currentPath),
			ParentWindowHandle: uintptr(windows.GetForegroundWindow()),
		})
	}, onPicked, onError)
}

func showExecutablePicker(_ fyne.Window, currentPath string, onPicked func(string), onError func(error)) {
	runNativeDialog(func() (string, error) {
		return cfdutil.ShowOpenFileDialog(cfd.DialogConfig{
			Title:               "Select Ikemen GO executable",
			DefaultFolder:       existingDialogDirectory(currentPath),
			ParentWindowHandle: uintptr(windows.GetForegroundWindow()),
			FileFilters: []cfd.FileFilter{
				{DisplayName: "Windows executable (*.exe)", Pattern: "*.exe"},
			},
		})
	}, onPicked, onError)
}

// runNativeDialog keeps COM initialization, dialog use, and release on one
// dedicated OS thread. go-common-file-dialog initializes COM internally but
// does not uninitialize it, so both successful initializations are balanced.
func runNativeDialog(run func() (string, error), onPicked func(string), onError func(error)) {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_DISABLE_OLE1DDE); err != nil {
			fyne.Do(func() { onError(fmt.Errorf("initialize Windows COM apartment: %w", err)) })
			return
		}
		defer ole.CoUninitialize()

		var path string
		var err error
		func() {
			// The helper calls CoInitializeEx again and releases the COM dialog
			// before returning. Balance that nested initialization afterwards.
			defer ole.CoUninitialize()
			path, err = run()
		}()

		if err != nil {
			if !errors.Is(err, cfd.ErrorCancelled) {
				fyne.Do(func() { onError(err) })
			}
			return
		}
		if path != "" {
			fyne.Do(func() { onPicked(path) })
		}
	}()
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
