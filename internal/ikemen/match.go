package ikemen

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// MatchConfig describes one CPU-versus-CPU fight launched through Ikemen GO.
type MatchConfig struct {
	InstallationDir string
	ExecutablePath  string
	Player1         string
	Player2         string
	AILevel         int
	RoundsToWin     int
	Stage           string
	GameSpeed       int
	GameSpeedMode   GameSpeedMode
}

// MatchResult contains the result parsed from Ikemen GO's text log.
type MatchResult struct {
	Stage       string
	Player1Wins int
	Player2Wins int
	Draws       int
	Winner      string
	IsDraw      bool
}

// RunMatch launches one fight and parses its log before returning.
// The executable is run with the installation directory as its working directory.
func RunMatch(ctx context.Context, config MatchConfig) (MatchResult, error) {
	if strings.TrimSpace(config.ExecutablePath) == "" {
		return MatchResult{}, fmt.Errorf("Ikemen GO executable path is required")
	}
	if strings.TrimSpace(config.Player1) == "" || strings.TrimSpace(config.Player2) == "" {
		return MatchResult{}, fmt.Errorf("both player identifiers are required")
	}
	if strings.EqualFold(config.Player1, config.Player2) {
		return MatchResult{}, fmt.Errorf("a character cannot fight itself")
	}
	if config.AILevel < 1 {
		return MatchResult{}, fmt.Errorf("AI level must be greater than 0")
	}
	if config.RoundsToWin < 1 {
		return MatchResult{}, fmt.Errorf("rounds to win must be greater than 0")
	}
	if err := ValidateGameSpeed(config.GameSpeedMode, config.GameSpeed); err != nil {
		return MatchResult{}, err
	}

	executable, err := filepath.Abs(config.ExecutablePath)
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not resolve Ikemen GO executable path: %w", err)
	}
	info, err := os.Stat(executable)
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not access Ikemen GO executable: %w", err)
	}
	if !info.Mode().IsRegular() {
		return MatchResult{}, fmt.Errorf("Ikemen GO executable is not a regular file")
	}

	workingDir := config.InstallationDir
	if strings.TrimSpace(workingDir) == "" {
		workingDir = filepath.Dir(executable)
	}
	workingDir, err = filepath.Abs(workingDir)
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not resolve installation directory: %w", err)
	}

	// Ikemen overwrites the selected log on each run. Give every invocation
	// its own temporary log file in the installation directory.
	tempLog, err := os.CreateTemp(workingDir, "ikemen-ranking-*.log")
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not create a temporary match log: %w", err)
	}
	logPath := tempLog.Name()
	logName := filepath.Base(logPath)
	if err := tempLog.Close(); err != nil {
		_ = os.Remove(logPath)
		return MatchResult{}, fmt.Errorf("could not prepare a temporary match log: %w", err)
	}
	if err := os.Remove(logPath); err != nil {
		return MatchResult{}, fmt.Errorf("could not prepare a temporary match log: %w", err)
	}
	defer os.Remove(logPath)

	stage := strings.TrimSpace(config.Stage)
	if stage == "" || strings.EqualFold(stage, RandomStage) {
		stages, loadErr := LoadStages(workingDir)
		if loadErr != nil {
			return MatchResult{}, fmt.Errorf("could not choose a random stage from data/select.def: %w", loadErr)
		}
		stage = stages[rand.Intn(len(stages))]
	}

	args := []string{
		config.Player1,
		config.Player2,
		"-p1.ai", strconv.Itoa(config.AILevel),
		"-p2.ai", strconv.Itoa(config.AILevel),
		"-rounds", strconv.Itoa(config.RoundsToWin),
		"-speed", strconv.Itoa(config.GameSpeed),
		"-windowed",
		"-log", logName,
	}
	args = append(args, "-s", stage)

	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = workingDir
	output, err := command.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return MatchResult{}, fmt.Errorf("Ikemen GO match was cancelled: %w", ctx.Err())
		}
		details := strings.TrimSpace(string(output))
		if details != "" {
			return MatchResult{}, fmt.Errorf("Ikemen GO exited with an error: %w: %s", err, details)
		}
		return MatchResult{}, fmt.Errorf("Ikemen GO exited with an error: %w", err)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not read Ikemen GO match log: %w", err)
	}
	result, err := ParseMatchLog(string(logData))
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not parse Ikemen GO match log: %w", err)
	}

	result.Stage = stage

	switch {
	case result.Player1Wins > result.Player2Wins:
		result.Winner = config.Player1
	case result.Player2Wins > result.Player1Wins:
		result.Winner = config.Player2
	default:
		result.IsDraw = true
	}
	return result, nil
}
