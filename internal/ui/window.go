package ui

import (
	"context"
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
	window.Resize(fyne.NewSize(1000, 760))

	title := widget.NewLabelWithStyle(
		"IkemenRanking",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	description := widget.NewLabel(
		"Automate CPU-vs-CPU fights in Ikemen GO and generate character rankings.",
	)
	installationPathLabel := widget.NewLabel("Installation: no folder selected.")
	executablePathLabel := widget.NewLabel("Executable: no file selected.")
	characterCount := widget.NewLabel("Characters found: 0")
	selectedCount := widget.NewLabel("Selected characters: 0")
	totalFights := widget.NewLabel("Total fights: 0")
	status := widget.NewLabel("Select the root folder of your Ikemen GO installation.")

	var installationDir string
	var executablePath string
	var matchCancel context.CancelFunc
	matchRunning := false

	characters := make([]string, 0)
	selected := make(map[string]bool)

	matchesPerMatchup := widget.NewEntry()
	matchesPerMatchup.SetText("1")
	matchesPerMatchup.SetPlaceHolder("1")

	player1Select := widget.NewSelect([]string{}, nil)
	player1Select.PlaceHolder = "Select player 1"
	player2Select := widget.NewSelect([]string{}, nil)
	player2Select.PlaceHolder = "Select player 2"

	aiLevel := widget.NewEntry()
	aiLevel.SetText("8")
	roundsToWin := widget.NewEntry()
	roundsToWin.SetText("2")
	stage := widget.NewEntry()
	stage.SetPlaceHolder("Optional stage identifier or path")

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

	updatePlayerOptions := func() {
		options := make([]string, 0, len(selected))
		for _, character := range characters {
			if selected[character] {
				options = append(options, character)
			}
		}

		oldPlayer1, oldPlayer2 := player1Select.Selected, player2Select.Selected
		player1Select.Options = options
		player2Select.Options = options
		if contains(options, oldPlayer1) {
			player1Select.SetSelected(oldPlayer1)
		} else {
			player1Select.SetSelected("")
		}
		if contains(options, oldPlayer2) {
			player2Select.SetSelected(oldPlayer2)
		} else {
			player2Select.SetSelected("")
		}
		player1Select.Refresh()
		player2Select.Refresh()
	}

	updateSelectedCount := func() {
		selectedCount.SetText(fmt.Sprintf("Selected characters: %d", len(selected)))
		updateTournamentSummary()
		updatePlayerOptions()
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

			// List rows are recycled while scrolling; always refresh both the
			// label and checkbox callback for the character at this row index.
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

		installationDir = path
		executablePath = ""
		executablePathLabel.SetText("Executable: no file selected.")
		characters = loaded
		selected = make(map[string]bool)
		installationPathLabel.SetText(fmt.Sprintf("Installation: %s", path))
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

	selectExecutable := widget.NewButton("Select executable", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if reader == nil {
				return
			}
			path := reader.URI().Path()
			if closeErr := reader.Close(); closeErr != nil {
				dialog.ShowError(closeErr, window)
				return
			}
			executablePath = path
			executablePathLabel.SetText(fmt.Sprintf("Executable: %s", path))
			status.SetText("Executable selected. Choose two participants to run a test match.")
		}, window)
		fileDialog.Show()
	})

	selectionControls := container.NewHBox(selectAll, deselectAll, selectedCount)
	tournamentControls := container.NewHBox(
		widget.NewLabel("Fights per matchup:"),
		matchesPerMatchup,
		totalFights,
	)

	runMatchButton := widget.NewButton("Run test match", nil)
	cancelMatchButton := widget.NewButton("Cancel match", func() {
		if matchCancel != nil {
			matchCancel()
			status.SetText("Cancelling match...")
		}
	})
	cancelMatchButton.Hide()

	runMatchButton.OnTapped = func() {
		if matchRunning {
			return
		}
		if installationDir == "" {
			dialog.ShowInformation("Installation required", "Select the root folder of your Ikemen GO installation first.", window)
			return
		}
		if executablePath == "" {
			dialog.ShowInformation("Executable required", "Select the Ikemen GO executable you want to run.", window)
			return
		}
		if player1Select.Selected == "" || player2Select.Selected == "" {
			dialog.ShowInformation("Participants required", "Select two participants for the test match.", window)
			return
		}
		if strings.EqualFold(player1Select.Selected, player2Select.Selected) {
			dialog.ShowInformation("Invalid participants", "Player 1 and Player 2 must be different characters.", window)
			return
		}
		ai, err := strconv.Atoi(strings.TrimSpace(aiLevel.Text))
		if err != nil || ai < 1 {
			dialog.ShowInformation("Invalid AI level", "AI level must be a whole number greater than 0.", window)
			return
		}
		rounds, err := strconv.Atoi(strings.TrimSpace(roundsToWin.Text))
		if err != nil || rounds < 1 {
			dialog.ShowInformation("Invalid rounds", "Rounds to win must be a whole number greater than 0.", window)
			return
		}

		config := ikemen.MatchConfig{
			InstallationDir: installationDir,
			ExecutablePath:  executablePath,
			Player1:         player1Select.Selected,
			Player2:         player2Select.Selected,
			AILevel:         ai,
			RoundsToWin:     rounds,
			Stage:           strings.TrimSpace(stage.Text),
		}

		ctx, cancel := context.WithCancel(context.Background())
		matchCancel = cancel
		matchRunning = true
		runMatchButton.Disable()
		cancelMatchButton.Show()
		status.SetText(fmt.Sprintf("Running test match: %s vs %s...", config.Player1, config.Player2))

		go func() {
			result, runErr := ikemen.RunMatch(ctx, config)
			fyne.Do(func() {
				defer cancel()
				matchCancel = nil
				matchRunning = false
				runMatchButton.Enable()
				cancelMatchButton.Hide()

				if runErr != nil {
					status.SetText("Test match failed.")
					dialog.ShowError(runErr, window)
					return
				}

				if result.IsDraw {
					status.SetText(fmt.Sprintf(
						"Test match finished: Draw. %s %d - %d %s (draws: %d).",
						config.Player1, result.Player1Wins, result.Player2Wins,
						config.Player2, result.Draws,
					))
					return
				}
				status.SetText(fmt.Sprintf(
					"Test match finished: %s won. %s %d - %d %s (draws: %d).",
					result.Winner, config.Player1, result.Player1Wins,
					result.Player2Wins, config.Player2, result.Draws,
				))
			})
		}()
	}

	header := container.NewVBox(
		title,
		widget.NewSeparator(),
		description,
		container.NewHBox(selectFolder, selectExecutable),
		installationPathLabel,
		executablePathLabel,
		characterCount,
		status,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Select the characters that will participate", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		selectionControls,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Tournament settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tournamentControls,
	)
	matchPanel := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Single match test", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, player1Select, player2Select),
		container.NewGridWithColumns(3,
			container.NewVBox(widget.NewLabel("AI level"), aiLevel),
			container.NewVBox(widget.NewLabel("Rounds to win"), roundsToWin),
			container.NewVBox(widget.NewLabel("Stage (optional)"), stage),
		),
		container.NewHBox(runMatchButton, cancelMatchButton),
	)

	window.SetContent(container.NewBorder(header, matchPanel, nil, nil, characterList))
	return window
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
