package ikemen

import (
	"context"
	"errors"
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
	RetryOnCrash    bool
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

// MaxCrashRetries is the number of additional attempts after Ikemen GO exits
// unsuccessfully. The initial attempt plus these retries make at most four runs.
const MaxCrashRetries = 3

type processExitError struct {
	details string
	logPath string
}

func (e *processExitError) Error() string {
	message := "Ikemen GO exited unexpectedly"
	if e.details != "" {
		message += ": " + e.details
	}
	return fmt.Sprintf("%s (match log: %s)", message, e.logPath)
}

// RunMatch launches one fight and parses its log before returning. When
// RetryOnCrash is enabled, only an unsuccessful process exit is retried;
// invalid configuration, cancellation, and match-log parsing errors are not.
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

	return runWithCrashRetries(config.RetryOnCrash, func() (MatchResult, error) {
		return runMatchAttempt(ctx, config)
	})
}

func runWithCrashRetries(retry bool, run func() (MatchResult, error)) (MatchResult, error) {
	for attempt := 0; ; attempt++ {
		result, err := run()
		if err == nil {
			return result, nil
		}
		var exitErr *processExitError
		if !retry || !errors.As(err, &exitErr) || attempt >= MaxCrashRetries {
			return MatchResult{}, err
		}
		// Intermediate crash logs are not needed once a retry is scheduled.
		_ = os.Remove(exitErr.logPath)
	}
}

func runMatchAttempt(ctx context.Context, config MatchConfig) (MatchResult, error) {
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
		return MatchResult{}, fmt.Errorf("could not prepare a temporary match log: %w", err)
	}
	if err := os.Remove(logPath); err != nil {
		return MatchResult{}, fmt.Errorf("could not prepare a temporary match log: %w", err)
	}

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
			_ = os.Remove(logPath)
			return MatchResult{}, fmt.Errorf("Ikemen GO match was cancelled: %w", ctx.Err())
		}
		details := strings.TrimSpace(string(output))
		return MatchResult{}, &processExitError{details: fmt.Sprintf("%v: %s", err, details), logPath: logPath}
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not read Ikemen GO match log at %s: %w", logPath, err)
	}
	result, err := ParseMatchLog(string(logData))
	if err != nil {
		return MatchResult{}, fmt.Errorf("could not parse Ikemen GO match log: %w (log preserved at %s)", err, logPath)
	}
	_ = os.Remove(logPath)

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
