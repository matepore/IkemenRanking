package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

// NewMainWindow builds the main application window.
func NewMainWindow(application fyne.App) fyne.Window {
	window := application.NewWindow("IkemenRanking")
	window.Resize(fyne.NewSize(900, 620))

	title := widget.NewLabelWithStyle(
		"IkemenRanking",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	description := widget.NewLabel(
		"Automate CPU-vs-CPU fights in Ikemen GO and generate character rankings.",
	)
	installationPath := widget.NewLabel("Installation: no folder selected.")
	characterCount := widget.NewLabel("Characters found: 0")
	selectedCount := widget.NewLabel("Selected characters: 0")
	totalFights := widget.NewLabel("Total fights: 0")
	status := widget.NewLabel("Select the root folder of your Ikemen GO installation.")

	characters := make([]string, 0)
	selected := make(map[string]bool)

	matchesPerMatchup := widget.NewEntry()
	matchesPerMatchup.SetText("1")
	matchesPerMatchup.SetPlaceHolder("1")

	updateTournamentSummary := func() {
		count, err := strconv.Atoi(strings.TrimSpace(matchesPerMatchup.Text))
		if err != nil || count < 1 {
			totalFights.SetText("Total fights: Enter a whole number greater than 0.")
			return
		}

		participantCount := len(selected)
		matchupCount := participantCount * (participantCount - 1) / 2
		totalFights.SetText(fmt.Sprintf("Total fights: %d", matchupCount*count))
	}

	updateSelectedCount := func() {
		selectedCount.SetText(fmt.Sprintf("Selected characters: %d", len(selected)))
		updateTournamentSummary()
	}

	matchesPerMatchup.OnChanged = func(string) {
		updateTournamentSummary()
	}

	characterList := widget.NewList(
		func() int { return len(characters) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel(""),
			)
		},
		func(item widget.ListItemID, object fyne.CanvasObject) {
			row := object.(*fyne.Container)
			check := row.Objects[0].(*widget.Check)
			name := row.Objects[1].(*widget.Label)
			character := characters[item]

			// Keep the checkbox and its label separate. List rows are recycled
			// while scrolling, so the label must be updated through SetText.
			name.SetText(character)
			check.OnChanged = func(checked bool) {
				if checked {
					selected[character] = true
				} else {
					delete(selected, character)
				}
				updateSelectedCount()
			}
			check.SetChecked(selected[character])
		},
	)

	loadInstallation := func(path string) {
		loaded, err := ikemen.LoadCharacters(path)
		if err != nil {
			status.SetText("Could not load the installation.")
			dialog.ShowError(err, window)
			return
		}

		characters = loaded
		selected = make(map[string]bool)
		installationPath.SetText(fmt.Sprintf("Installation: %s", path))
		characterCount.SetText(fmt.Sprintf("Characters found: %d", len(characters)))
		updateSelectedCount()

		if len(characters) == 0 {
			status.SetText("data/select.def was read, but [Characters] contains no entries.")
		} else {
			status.SetText("Character list loaded. Select the participants for the tournament.")
		}
		characterList.Refresh()
	}

	selectAll := widget.NewButton("Select all", func() {
		for _, character := range characters {
			selected[character] = true
		}
		updateSelectedCount()
		characterList.Refresh()
	})

	deselectAll := widget.NewButton("Deselect all", func() {
		selected = make(map[string]bool)
		updateSelectedCount()
		characterList.Refresh()
	})

	selectFolder := widget.NewButton("Select Ikemen GO folder", func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri == nil {
				return
			}
			loadInstallation(uri.Path())
		}, window)
		folderDialog.Show()
	})

	selectionControls := container.NewHBox(selectAll, deselectAll, selectedCount)
	tournamentControls := container.NewHBox(
		widget.NewLabel("Fights per matchup:"),
		matchesPerMatchup,
		totalFights,
	)
	header := container.NewVBox(
		title,
		widget.NewSeparator(),
		description,
		selectFolder,
		installationPath,
		characterCount,
		status,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Select the characters that will participate", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		selectionControls,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Tournament settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tournamentControls,
	)

	window.SetContent(container.NewBorder(header, nil, nil, nil, characterList))

	return window
}
