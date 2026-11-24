package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
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

	description := widget.NewLabel(
		"Automate CPU-vs-CPU fights in Ikemen GO and generate character rankings.",
	)
	installationPathEntry := newPathEntry("Path to Ikemen GO folder")
	executablePathEntry := newPathEntry("Path to Ikemen GO executable")
	installationPathStatus := widget.NewLabel("No installation folder selected.")
	executablePathStatus := widget.NewLabel("No executable selected.")
	characterCount := widget.NewLabel("Characters found: 0")
	selectedCount := widget.NewLabel("Selected characters: 0")
	totalFights := widget.NewLabel("Total fights: 0")
	status := widget.NewLabel("Select the root folder of your Ikemen GO installation.")
	tournamentProgressLabel := widget.NewLabel("Tournament progress: not started.")
	rankingSummary := widget.NewLabel("Final ranking will appear here.")
	tournamentProgress := widget.NewProgressBar()

	var installationDir string
	var executablePath string
	var tournamentCancel context.CancelFunc
	var tournamentPause *tournament.PauseController
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
	retryCrashes := widget.NewCheck("Retry fights if Ikemen GO crashes unexpectedly (up to 3 retries)", nil)
	var gameSpeedMode ikemen.GameSpeedMode
	gameSpeedModeDetected := false
	gameSpeedValues := map[string]int{}
	gameSpeed := widget.NewSelect(nil, nil)
	gameSpeed.Disable()
	gameSpeedInfo := widget.NewLabel("Game speed system: select an Ikemen GO installation to detect compatibility.")
	updateGameSpeedOptions := func(mode ikemen.GameSpeedMode) error {
		options, values, defaultValue, err := ikemen.GameSpeedOptions(mode)
		if err != nil { return err }
		gameSpeedMode = mode
		gameSpeedValues = values
		gameSpeed.Options = options
		gameSpeed.SetSelected(defaultValue)
		gameSpeed.Enable()
		gameSpeedModeDetected = true
		gameSpeedInfo.SetText("Detected game speed system: " + mode.String())
		return nil
	}

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

	loadInstallation := func(path string) error {
		loaded, err := ikemen.LoadCharacters(path)
		if err != nil {
			return err
		}

		installationDir = path
		installationPathStatus.SetText("✓ Valid Ikemen GO folder.")
		installationPathEntry.SetText(path)
		gameSpeedModeDetected = false
		gameSpeed.Disable()
		gameSpeedInfo.SetText("Detecting Ikemen GO game speed system...")
		if detectedMode, detectErr := ikemen.DetectGameSpeedMode(path); detectErr != nil {
			gameSpeedInfo.SetText("Game speed system: detection failed. Select the Ikemen GO root folder.")
		} else if err := updateGameSpeedOptions(detectedMode); err != nil {
			gameSpeedInfo.SetText("Game speed system: detection failed.")
		}
		executablePath = ""
		executablePathEntry.SetText("")
		executablePathStatus.SetText("No executable selected.")
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
		return nil
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

	clearInvalidInstallation := func(message string) {
		installationDir = ""
		gameSpeedModeDetected = false
		gameSpeed.Disable()
		gameSpeedInfo.SetText("Game speed system: select an Ikemen GO installation to detect compatibility.")
		characters = nil
		selected = make(map[string]bool)
		characterCount.SetText("Characters found: 0")
		stage.Options = []string{ikemen.RandomStage}
		stagePaths = map[string]string{ikemen.RandomStage: ikemen.RandomStage}
		stage.SetSelected(ikemen.RandomStage)
		updateSelectedCount()
		characterList.Refresh()
		installationPathStatus.SetText(message)
	}

	validateInstallationPath := func() {
		path := strings.TrimSpace(installationPathEntry.Text)
		if path == "" {
			clearInvalidInstallation("⚠ Invalid folder path. Enter an Ikemen GO installation folder.")
			return
		}
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			clearInvalidInstallation("⚠ Invalid folder path. The folder does not exist or is not accessible.")
			return
		}
		if installationDir != "" && filepath.Clean(path) == filepath.Clean(installationDir) {
			installationPathStatus.SetText("✓ Valid Ikemen GO folder.")
			return
		}
		if err := loadInstallation(path); err != nil {
			clearInvalidInstallation("⚠ Invalid Ikemen GO folder: " + err.Error())
			status.SetText("Could not load the selected Ikemen GO installation.")
		}
	}

	validateExecutablePath := func() {
		path := strings.TrimSpace(executablePathEntry.Text)
		info, err := os.Stat(path)
		if path == "" || err != nil || !info.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(path), ".exe") {
			executablePath = ""
			executablePathStatus.SetText("⚠ Invalid executable path. Select an existing .exe file.")
			return
		}
		executablePath = path
		executablePathStatus.SetText("✓ Valid executable.")
	}

	installationPathEntry.OnSubmitted = func(string) { validateInstallationPath() }
	installationPathEntry.onFocusLost = validateInstallationPath
	executablePathEntry.OnSubmitted = func(string) { validateExecutablePath() }
	executablePathEntry.onFocusLost = validateExecutablePath

	browseFolder := widget.NewButton("Browse", func() {
		showFolderPicker(window, installationPathEntry.Text, func(path string) {
			installationPathEntry.SetText(path)
			validateInstallationPath()
		}, func(err error) { dialog.ShowError(err, window) })
	})
	browseExecutable := widget.NewButton("Browse", func() {
		showExecutablePicker(window, executablePathEntry.Text, func(path string) {
			executablePathEntry.SetText(path)
			validateExecutablePath()
		}, func(err error) { dialog.ShowError(err, window) })
	})

	selectionControls := container.NewHBox(selectAll, deselectAll, selectedCount)
	tournamentControls := container.NewVBox(totalFights, replayPreviouslyCompleted, retryCrashes)

	header := newSectionCard(
		"IkemenRanking",
		"CPU-vs-CPU tournament manager",
		container.NewVBox(
			description,
			widget.NewLabel("Ikemen GO folder"),
			container.NewBorder(nil, nil, nil, browseFolder, installationPathEntry),
			installationPathStatus,
			widget.NewLabel("Executable"),
			container.NewBorder(nil, nil, nil, browseExecutable, executablePathEntry),
			executablePathStatus,
			status,
		),
	)
	selectionPanel := newSectionCard(
		"Character selection",
		"Choose the participants for the round-robin tournament",
		selectionControls,
	)
	tournamentSettingsPanel := newSectionCard(
		"Tournament settings",
		"Current selection overview",
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
		if tournamentRunning {
			return
		}
		validateInstallationPath()
		validateExecutablePath()
		if installationDir == "" {
			dialog.ShowInformation("Installation required", "Select the root folder of your Ikemen GO installation first.", window)
			return
		}
		if executablePath == "" {
			dialog.ShowInformation("Executable required", "Select the Ikemen GO executable you want to run.", window)
			return
		}
		if !gameSpeedModeDetected {
			dialog.ShowInformation("Unsupported Ikemen GO version", "IkemenRanking could not identify this installation's -speed implementation. Make sure you selected the root folder of the Ikemen GO installation.", window)
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
			GameSpeedMode: gameSpeedMode,
			RetryOnCrash: retryCrashes.Checked,
		}
		replay := replayPreviouslyCompleted.Checked
		ctx, cancel := context.WithCancel(context.Background())
		tournamentCancel = cancel
		pauseController := &tournament.PauseController{}
		tournamentPause = pauseController
		tournamentRunning = true
		tournamentRunButton.Disable()
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

	matchPanel := newSectionCard(
		"Match configuration",
		"Configure the fights played during the tournament",
		container.NewVBox(
		widget.NewLabel("Choose Random to select a stage at random."),
		container.NewGridWithColumns(4,
			container.NewVBox(widget.NewLabel("AI level"), aiLevel),
			container.NewVBox(widget.NewLabel("Rounds to win"), roundsToWin),
			container.NewVBox(widget.NewLabel("Stage"), stage),
			container.NewVBox(widget.NewLabel("Game speed"), gameSpeed),
		),
		gameSpeedInfo,
		widget.NewLabel("Only normal and faster speeds are available. Options depend on the detected Ikemen GO version."),
		),
	)

	tournamentPanel := newSectionCard(
		"Tournament execution",
		"Track progress, pause or cancel the current run",
		container.NewVBox(
		tournamentProgressLabel,
		tournamentProgress,
		container.NewHBox(tournamentRunButton, pauseTournamentButton, resumeTournamentButton, cancelTournamentButton),
		rankingSummary,
		),
	)
	characterListArea := newSectionCard(
		"Character roster",
		"Select characters from the loaded Ikemen GO installation",
		container.NewVBox(
			characterCount,
			container.NewGridWrap(fyne.NewSize(1000, 360), characterList),
		),
	)
	pageContent := container.NewVBox(
		header,
		selectionPanel,
		characterListArea,
		matchPanel,
		tournamentSettingsPanel,
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
