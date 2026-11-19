package ui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	pathpkg "path"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/matepore/IkemenRanking/internal/ikemen"
	"github.com/matepore/IkemenRanking/internal/ranking"
	"github.com/matepore/IkemenRanking/internal/storage"
	"github.com/matepore/IkemenRanking/internal/tournament"
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
	tournamentProgressLabel := widget.NewLabel("Tournament progress: not started.")
	rankingSummary := widget.NewLabel("Final ranking will appear here.")
	tournamentProgress := widget.NewProgressBar()

	var installationDir string
	var executablePath string
	var matchCancel context.CancelFunc
	var tournamentCancel context.CancelFunc
	var tournamentPause *tournament.PauseController
	matchRunning := false
	tournamentRunning := false

	characters := make([]string, 0)
	selected := make(map[string]bool)

	aiLevel := widget.NewEntry()
	aiLevel.SetText("8")
	roundsToWin := widget.NewEntry()
	roundsToWin.SetText("2")
	stage := widget.NewSelect([]string{ikemen.RandomStage}, nil)
	stage.SetSelected(ikemen.RandomStage)
	stagePaths := map[string]string{ikemen.RandomStage: ikemen.RandomStage}
	replayPreviouslyCompleted := widget.NewCheck("Replay previously completed fights", nil)
	gameSpeedOptions := []string{"Slower (-9)", "Slower (-8)", "Slower (-7)", "Slower (-6)", "Slower (-5)", "Slower (-4)", "Slower (-3)", "Slower (-2)", "Slower (-1)", "Normal", "Faster (+1)", "Faster (+2)", "Faster (+3)", "Faster (+4)", "Faster (+5)", "Faster (+6)", "Faster (+7)", "Faster (+8)", "Faster (+9)"}
	gameSpeedValues := map[string]int{
		"Slower (-9)": -9, "Slower (-8)": -8, "Slower (-7)": -7, "Slower (-6)": -6,
		"Slower (-5)": -5, "Slower (-4)": -4, "Slower (-3)": -3, "Slower (-2)": -2,
		"Slower (-1)": -1, "Normal": 0, "Faster (+1)": 1, "Faster (+2)": 2,
		"Faster (+3)": 3, "Faster (+4)": 4, "Faster (+5)": 5, "Faster (+6)": 6,
		"Faster (+7)": 7, "Faster (+8)": 8, "Faster (+9)": 9,
	}
	gameSpeed := widget.NewSelect(gameSpeedOptions, nil)
	gameSpeed.SetSelected("Normal")

	updateTournamentSummary := func() {
		participantCount := len(selected)
		matchupCount := participantCount * (participantCount - 1) / 2
		totalFights.SetText(fmt.Sprintf("Total fights: %d", matchupCount))
	}

	updateSelectedCount := func() {
		selectedCount.SetText(fmt.Sprintf("Selected characters: %d", len(selected)))
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
		stages, stagesErr := ikemen.LoadStages(path)
		stageOptions := []string{ikemen.RandomStage}
		stagePaths = map[string]string{ikemen.RandomStage: ikemen.RandomStage}
		if stagesErr == nil {
			usedLabels := map[string]int{}
			for _, stagePath := range stages {
				label := stagePath
				if len(label) >= len("stages/") {
					label = label[len("stages/"):]
				}
				if strings.EqualFold(pathpkg.Ext(label), ".def") {
					label = label[:len(label)-len(pathpkg.Ext(label))]
				}
				baseLabel := label
				usedLabels[strings.ToLower(baseLabel)]++
				if usedLabels[strings.ToLower(baseLabel)] > 1 {
					label = fmt.Sprintf("%s (%d)", baseLabel, usedLabels[strings.ToLower(baseLabel)])
				}
				stageOptions = append(stageOptions, label)
				stagePaths[label] = stagePath
			}
		}
		stage.Options = stageOptions
		stage.SetSelected(ikemen.RandomStage)
		selected = make(map[string]bool)
		installationPathLabel.SetText(fmt.Sprintf("Installation: %s", path))
		characterCount.SetText(fmt.Sprintf("Characters found: %d", len(characters)))
		updateSelectedCount()

		if len(characters) == 0 {
			status.SetText("data/select.def was read, but [Characters] contains no entries.")
		} else if stagesErr != nil {
			status.SetText("Characters loaded, but stages could not be read from [ExtraStages].")
			dialog.ShowError(stagesErr, window)
		} else {
			status.SetText(fmt.Sprintf("Loaded %d characters and %d stages. Select the tournament participants.", len(characters), len(stages)))
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
			status.SetText("Executable selected. Select at least two participants to run a test match.")
		}, window)
		fileDialog.Show()
	})

	selectionControls := container.NewHBox(selectAll, deselectAll, selectedCount)
	tournamentControls := container.NewVBox(totalFights, replayPreviouslyCompleted)

	runMatchButton := widget.NewButton("Run test match", nil)
	cancelMatchButton := widget.NewButton("Cancel match", func() {
		if matchCancel != nil {
			matchCancel()
			status.SetText("Cancelling match...")
		}
	})
	cancelMatchButton.Hide()

	runMatchButton.OnTapped = func() {
		if matchRunning || tournamentRunning {
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
		participants := make([]string, 0, 2)
		for _, character := range characters {
			if selected[character] {
				participants = append(participants, character)
				if len(participants) == 2 {
					break
				}
			}
		}
		if len(participants) < 2 {
			dialog.ShowInformation("Participants required", "Select at least two participants for the test match.", window)
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
			Player1:         participants[0],
			Player2:         participants[1],
			AILevel:         ai,
			RoundsToWin:     rounds,
			Stage:           stagePaths[stage.Selected],
			GameSpeed:       gameSpeedValues[gameSpeed.Selected],
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
	tournamentRunButton := widget.NewButton("Run tournament", nil)
	cancelTournamentButton := widget.NewButton("Cancel tournament", func() {
		if tournamentCancel != nil {
			tournamentCancel()
			status.SetText("Cancelling tournament...")
		}
	})
	cancelTournamentButton.Hide()
	var pauseTournamentButton *widget.Button
	var resumeTournamentButton *widget.Button
	pauseTournamentButton = widget.NewButton("Pause tournament", func() {
		if tournamentPause != nil {
			tournamentPause.Pause()
			pauseTournamentButton.Hide()
			resumeTournamentButton.Show()
			resumeTournamentButton.Disable()
			status.SetText("Pause requested. The current fight will finish before the tournament pauses.")
		}
	})
	resumeTournamentButton = widget.NewButton("Resume tournament", func() {
		if tournamentPause != nil {
			tournamentPause.Resume()
			resumeTournamentButton.Hide()
			pauseTournamentButton.Show()
			status.SetText("Resuming tournament...")
		}
	})
	pauseTournamentButton.Hide()
	resumeTournamentButton.Hide()
	resumeTournamentButton.Disable()

	tournamentRunButton.OnTapped = func() {
		if matchRunning || tournamentRunning {
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

		participants := make([]string, 0, len(selected))
		for _, character := range characters {
			if selected[character] {
				participants = append(participants, character)
			}
		}
		if len(participants) < 2 {
			dialog.ShowInformation("Participants required", "Select at least two participants for the tournament.", window)
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
		plannedMatches, err := tournament.GenerateMatches(participants)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		totalFightCount := len(plannedMatches)
		outputDirectory, err := storage.EnsureResultsDir()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		database, err := storage.OpenResultsDatabase(outputDirectory)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		config := tournament.Config{
			InstallationDir: installationDir,
			ExecutablePath: executablePath,
			Participants: participants,
			AILevel: ai,
			RoundsToWin: rounds,
			Stage: stagePaths[stage.Selected],
			GameSpeed: gameSpeedValues[gameSpeed.Selected],
		}
		replay := replayPreviouslyCompleted.Checked
		ctx, cancel := context.WithCancel(context.Background())
		tournamentCancel = cancel
		pauseController := &tournament.PauseController{}
		tournamentPause = pauseController
		tournamentRunning = true
		tournamentRunButton.Disable()
		runMatchButton.Disable()
		cancelTournamentButton.Show()
		pauseTournamentButton.Show()
		resumeTournamentButton.Hide()
		resumeTournamentButton.Disable()
		tournamentProgress.SetValue(0)
		tournamentProgressLabel.SetText("Tournament progress: starting...")
		rankingSummary.SetText("Tournament is running.")
		status.SetText("Tournament started.")

		go func() {
			defer database.Close()
			lastProgress := tournament.Progress{Total: totalFightCount}
			var progressExportErr error
			records, runErr := tournament.RunPersistentWithControl(ctx, config, database, replay, ikemen.RunMatch, func(progress tournament.Progress) {
				lastProgress = progress
				// The fight has already been persisted when this callback runs.
				// Refresh the global report immediately so it stays current during a tournament.
				if _, err := exportGlobalRanking(context.Background(), database, outputDirectory); err != nil {
					progressExportErr = errors.Join(progressExportErr, err)
				}
				fyne.Do(func() {
					if progress.Total > 0 {
						tournamentProgress.SetValue(float64(progress.Completed) / float64(progress.Total))
					}
					tournamentProgressLabel.SetText(fmt.Sprintf(
						"Tournament progress: %d/%d processed — played %d, skipped %d. %s vs %s",
						progress.Completed, progress.Total, progress.Played, progress.Skipped,
						progress.Match.Player1, progress.Match.Player2,
					))
				})
			}, pauseController, func() {
				fyne.Do(func() {
					status.SetText("Tournament paused. Resume to continue.")
					resumeTournamentButton.Enable()
				})
			})

			// Final refresh also covers runs in which every matchup was skipped.
			standings, finalExportErr := exportGlobalRanking(context.Background(), database, outputDirectory)
			exportErr := errors.Join(progressExportErr, finalExportErr)

			fyne.Do(func() {
				defer cancel()
				tournamentCancel = nil
				tournamentPause = nil
				tournamentRunning = false
				tournamentRunButton.Enable()
				runMatchButton.Enable()
				cancelTournamentButton.Hide()
				pauseTournamentButton.Hide()
				resumeTournamentButton.Hide()

				var summary strings.Builder
				summary.WriteString("Position  Character  Points  W-L-D\n")
				for _, standing := range standings {
					summary.WriteString(fmt.Sprintf(
						"%d. %s  %d  %d-%d-%d\n",
						standing.Position, standing.Character, standing.Points,
						standing.Wins, standing.Losses, standing.Draws,
					))
				}
				rankingSummary.SetText(summary.String())
				if totalFightCount > 0 {
					tournamentProgress.SetValue(float64(len(records)) / float64(totalFightCount))
				}
				tournamentProgressLabel.SetText(fmt.Sprintf(
					"Tournament progress: %d/%d matchups have saved results. Played %d, skipped %d.",
					len(records), totalFightCount, lastProgress.Played, lastProgress.Skipped,
				))

				completionMessage := ""
				if runErr != nil {
					completionMessage = fmt.Sprintf(
						"Tournament stopped with %d/%d saved results. Played %d, skipped %d.",
						len(records), totalFightCount, lastProgress.Played, lastProgress.Skipped,
					)
				} else {
					completionMessage = fmt.Sprintf(
						"Tournament complete. Played %d fights and skipped %d previously completed matchups.",
						lastProgress.Played, lastProgress.Skipped,
					)
				}
				if exportErr != nil {
					status.SetText(completionMessage)
					dialog.ShowError(fmt.Errorf("tournament data could not be fully exported: %w", exportErr), window)
				} else {
					status.SetText(fmt.Sprintf("%s HTML report and CSV files saved to %s.", completionMessage, outputDirectory))
				}
				if runErr != nil {
					dialog.ShowError(runErr, window)
				}
			})
		}()
	}

	matchPanel := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Match settings and single-match test", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("The first two selected participants (in select.def order) will be used."),
		widget.NewLabel("Choose Random to select a stage at random."),
		container.NewGridWithColumns(4,
			container.NewVBox(widget.NewLabel("AI level"), aiLevel),
			container.NewVBox(widget.NewLabel("Rounds to win"), roundsToWin),
			container.NewVBox(widget.NewLabel("Stage"), stage),
			container.NewVBox(widget.NewLabel("Game speed"), gameSpeed),
		),
		widget.NewLabel("Game speed follows Ikemen GO’s -speed option: 0 is normal, positive values are faster, and negative values are slower."),
		container.NewHBox(runMatchButton, cancelMatchButton),
	)

	tournamentPanel := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Tournament execution", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tournamentProgressLabel,
		tournamentProgress,
		container.NewHBox(tournamentRunButton, pauseTournamentButton, resumeTournamentButton, cancelTournamentButton),
		rankingSummary,
	)
	characterListArea := container.NewGridWrap(fyne.NewSize(1000, 360), characterList)
	pageContent := container.NewVBox(
		header,
		characterListArea,
		matchPanel,
		tournamentPanel,
	)
	window.SetContent(container.NewVScroll(pageContent))
	return window
}


 // exportGlobalRanking regenerates all reports from every current saved matchup.
 // It is called after each completed fight and once again when a run finishes.
func exportGlobalRanking(ctx context.Context, database *storage.Database, outputDirectory string) ([]ranking.Standing, error) {
	records, err := database.ListAllFights(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string)
	for _, record := range records {
		for _, name := range []string{record.Match.Player1, record.Match.Player2} {
			key := strings.ToLower(name)
			if _, exists := names[key]; !exists {
				names[key] = name
			}
		}
	}
	participants := make([]string, 0, len(names))
	for _, name := range names {
		participants = append(participants, name)
	}
	sort.Strings(participants)

	outcomes := make([]ranking.MatchOutcome, 0, len(records))
	for _, record := range records {
		outcomes = append(outcomes, ranking.MatchOutcome{
			Player1: record.Match.Player1,
			Player2: record.Match.Player2,
			Winner: record.Result.Winner,
			IsDraw: record.Result.IsDraw,
		})
	}
	standings, err := ranking.Calculate(participants, outcomes)
	if err != nil {
		return nil, err
	}

	var exportErr error
	if err := storage.ExportMatchResultsCSV(filepath.Join(outputDirectory, "ikemen-ranking-matches.csv"), records); err != nil {
		exportErr = errors.Join(exportErr, err)
	}
	if err := storage.ExportStandingsCSV(filepath.Join(outputDirectory, "ikemen-ranking-standings.csv"), standings); err != nil {
		exportErr = errors.Join(exportErr, err)
	}
	if err := storage.ExportRankingHTML(filepath.Join(outputDirectory, "ikemen-ranking-report.html"), storage.RankingReport{
		GeneratedAt: time.Now(),
		Standings: standings,
		Matches: records,
	}); err != nil {
		exportErr = errors.Join(exportErr, err)
	}
	return standings, exportErr
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
