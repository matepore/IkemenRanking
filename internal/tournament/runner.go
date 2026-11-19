package tournament

import (
	"context"
	"fmt"
	"strings"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

type Config struct {
	InstallationDir string
	ExecutablePath  string
	Participants    []string
	AILevel         int
	RoundsToWin     int
	Stage           string
	GameSpeed       int
}

type MatchRecord struct {
	Index  int
	Match  Match
	Result ikemen.MatchResult
}

type MatchRunner func(context.Context, ikemen.MatchConfig) (ikemen.MatchResult, error)

type FightStore interface {
	ListFightsForParticipants(context.Context, []string) ([]MatchRecord, error)
	UpsertFight(context.Context, MatchRecord) error
}

type Progress struct {
	Completed int
	Total     int
	Played    int
	Skipped   int
	Match     Match
}

func Run(ctx context.Context, config Config, runMatch MatchRunner, onProgress func(Progress)) ([]MatchRecord, error) {
	return run(ctx, config, nil, true, runMatch, onProgress, nil, nil)
}

func RunPersistent(ctx context.Context, config Config, store FightStore, replay bool, runMatch MatchRunner, onProgress func(Progress)) ([]MatchRecord, error) {
	return RunPersistentWithControl(ctx, config, store, replay, runMatch, onProgress, nil, nil)
}

// RunPersistentWithControl executes pending fights, persists each successful
// result immediately, and optionally pauses safely between fights.
func RunPersistentWithControl(
	ctx context.Context,
	config Config,
	store FightStore,
	replay bool,
	runMatch MatchRunner,
	onProgress func(Progress),
	pause *PauseController,
	onPaused func(),
) ([]MatchRecord, error) {
	if store == nil {
		return nil, fmt.Errorf("fight store is required")
	}
	return run(ctx, config, store, replay, runMatch, onProgress, pause, onPaused)
}

func run(
	ctx context.Context,
	config Config,
	store FightStore,
	replay bool,
	runMatch MatchRunner,
	onProgress func(Progress),
	pause *PauseController,
	onPaused func(),
) ([]MatchRecord, error) {
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
	if config.GameSpeed < -9 || config.GameSpeed > 9 {
		return nil, fmt.Errorf("game speed must be between -9 and 9")
}

	saved := make(map[string]MatchRecord, len(matches))
	if store != nil {
		existing, err := store.ListFightsForParticipants(ctx, config.Participants)
		if err != nil {
			return nil, fmt.Errorf("could not load saved fight results: %w", err)
		}
		for _, record := range existing {
			key, err := matchupKey(record.Match.Player1, record.Match.Player2)
			if err != nil {
				return nil, fmt.Errorf("invalid saved fight: %w", err)
			}
			saved[key] = record
		}
	}

	processed, played, skipped := 0, 0, 0
	for index, match := range matches {
		if pause != nil {
			if err := pause.Wait(ctx, onPaused); err != nil {
				return collectRecords(matches, saved), err
			}
		}
		if err := ctx.Err(); err != nil {
			return collectRecords(matches, saved), err
		}
		key, _ := matchupKey(match.Player1, match.Player2)
		if existing, exists := saved[key]; exists && !replay {
			saved[key] = orientRecord(existing, match, index+1)
			processed++
			skipped++
			if onProgress != nil {
				onProgress(Progress{Completed: processed, Total: len(matches), Played: played, Skipped: skipped, Match: match})
			}
			continue
		}

		result, runErr := runMatch(ctx, ikemen.MatchConfig{
			InstallationDir: config.InstallationDir,
			ExecutablePath:  config.ExecutablePath,
			Player1:         match.Player1,
			Player2:         match.Player2,
			AILevel:         config.AILevel,
			RoundsToWin:     config.RoundsToWin,
			Stage:           config.Stage,
			GameSpeed:       config.GameSpeed,
		})
		if runErr != nil {
			return collectRecords(matches, saved), fmt.Errorf("fight %d of %d (%s vs %s) failed: %w",
				index+1, len(matches), match.Player1, match.Player2, runErr)
		}
		record := MatchRecord{Index: index + 1, Match: match, Result: result}
		if store != nil {
			if err := store.UpsertFight(ctx, record); err != nil {
				return collectRecords(matches, saved), fmt.Errorf("could not persist fight %d of %d (%s vs %s): %w",
					index+1, len(matches), match.Player1, match.Player2, err)
			}
		}
		saved[key] = record
		processed++
		played++
		if onProgress != nil {
			onProgress(Progress{Completed: processed, Total: len(matches), Played: played, Skipped: skipped, Match: match})
		}
	}
	return collectRecords(matches, saved), nil
}

func matchupKey(player1, player2 string) (string, error) {
	key1, key2 := strings.ToLower(strings.TrimSpace(player1)), strings.ToLower(strings.TrimSpace(player2))
	if key1 == "" || key2 == "" {
		return "", fmt.Errorf("participant names cannot be empty")
	}
	if key1 == key2 {
		return "", fmt.Errorf("a character cannot fight itself")
	}
	if key1 > key2 {
		key1, key2 = key2, key1
	}
	return key1 + "\x00" + key2, nil
}

func orientRecord(record MatchRecord, match Match, index int) MatchRecord {
	if !strings.EqualFold(record.Match.Player1, match.Player1) {
		record.Result.Player1Wins, record.Result.Player2Wins = record.Result.Player2Wins, record.Result.Player1Wins
	}
	record.Index = index
	record.Match = match
	return record
}

func collectRecords(matches []Match, saved map[string]MatchRecord) []MatchRecord {
	records := make([]MatchRecord, 0, len(saved))
	for index, match := range matches {
		key, _ := matchupKey(match.Player1, match.Player2)
		if record, exists := saved[key]; exists {
			records = append(records, orientRecord(record, match, index+1))
		}
	}
	return records
}
