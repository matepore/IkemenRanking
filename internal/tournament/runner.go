package tournament

import (
	"context"
	"fmt"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

// Config contains the settings shared by every fight in a tournament.
type Config struct {
	InstallationDir string
	ExecutablePath  string
	Participants    []string
	AILevel         int
	RoundsToWin     int
	Stage           string
}

// MatchRecord stores one completed fight and the result reported by Ikemen GO.
type MatchRecord struct {
	Index  int
	Match  Match
	Result ikemen.MatchResult
}

// MatchRunner executes one Ikemen GO fight.
type MatchRunner func(context.Context, ikemen.MatchConfig) (ikemen.MatchResult, error)

// Progress describes the most recently completed fight.
type Progress struct {
	Completed int
	Total     int
	Match     Match
}

// Run executes the generated fights sequentially. It returns all completed
// records if a fight fails or the context is cancelled, so callers can retain
// partial progress.
func Run(ctx context.Context, config Config, runMatch MatchRunner, onProgress func(Progress)) ([]MatchRecord, error) {
	if runMatch == nil {
		return nil, fmt.Errorf("match runner is required")
	}
	matches, err := GenerateMatches(config.Participants)
	if err != nil {
		return nil, err
	}
	if config.AILevel < 1 {
		return nil, fmt.Errorf("AI level must be greater than 0")
	}
	if config.RoundsToWin < 1 {
		return nil, fmt.Errorf("rounds to win must be greater than 0")
	}

	records := make([]MatchRecord, 0, len(matches))
	for index, match := range matches {
		if err := ctx.Err(); err != nil {
			return records, err
		}
		result, runErr := runMatch(ctx, ikemen.MatchConfig{
			InstallationDir: config.InstallationDir,
			ExecutablePath:  config.ExecutablePath,
			Player1:         match.Player1,
			Player2:         match.Player2,
			AILevel:         config.AILevel,
			RoundsToWin:     config.RoundsToWin,
			Stage:           config.Stage,
		})
		if runErr != nil {
			return records, fmt.Errorf("fight %d of %d (%s vs %s) failed: %w",
				index+1, len(matches), match.Player1, match.Player2, runErr)
		}
		records = append(records, MatchRecord{
			Index:  index + 1,
			Match:  match,
			Result: result,
		})
		if onProgress != nil {
			onProgress(Progress{
				Completed: len(records),
				Total:     len(matches),
				Match:     match,
			})
		}
	}
	return records, nil
}
