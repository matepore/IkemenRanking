package tournament

import (
	"context"
	"errors"
	"testing"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

func TestRunExecutesMatchesSequentiallyAndReportsProgress(t *testing.T) {
	config := Config{
		Participants: []string{"A", "B", "C"},
		AILevel: 8,
		RoundsToWin: 2,
		Stage: "Random",
		GameSpeed: 4,
		GameSpeedMode: ikemen.GameSpeedModeSteps,
	}
	calls := 0
	progress := make([]Progress, 0)
	records, err := Run(context.Background(), config, func(_ context.Context, matchConfig ikemen.MatchConfig) (ikemen.MatchResult, error) {
		calls++
		if matchConfig.AILevel != 8 || matchConfig.RoundsToWin != 2 || matchConfig.Stage != "Random" || matchConfig.GameSpeed != 4 || matchConfig.GameSpeedMode != ikemen.GameSpeedModeSteps {
			t.Fatalf("unexpected match config: %#v", matchConfig)
		}
		return ikemen.MatchResult{Winner: matchConfig.Player1, Player1Wins: 2}, nil
	}, func(update Progress) {
		progress = append(progress, update)
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if calls != 3 || len(records) != 3 || len(progress) != 3 {
		t.Fatalf("calls=%d records=%d progress=%d, want 3 each", calls, len(records), len(progress))
	}
	if progress[2].Completed != 3 || progress[2].Total != 3 {
		t.Fatalf("unexpected final progress: %#v", progress[2])
	}
}

func TestRunReturnsCompletedRecordsWhenMatchFails(t *testing.T) {
	wantErr := errors.New("game failed")
	config := Config{Participants: []string{"A", "B", "C"}, AILevel: 8, RoundsToWin: 2}
	calls := 0
	records, err := Run(context.Background(), config, func(context.Context, ikemen.MatchConfig) (ikemen.MatchResult, error) {
		calls++
		if calls == 2 {
			return ikemen.MatchResult{}, wantErr
		}
		return ikemen.MatchResult{Winner: "A"}, nil
	}, nil)
	if err == nil {
		t.Fatal("Run() error = nil, want failure")
	}
	if len(records) != 1 {
		t.Fatalf("got %d completed records, want 1", len(records))
	}
}

func TestRunRejectsInvalidSettings(t *testing.T) {
	config := Config{Participants: []string{"A", "B"}, AILevel: 0, RoundsToWin: 2}
	if _, err := Run(context.Background(), config, func(context.Context, ikemen.MatchConfig) (ikemen.MatchResult, error) {
		t.Fatal("runner should not be called for invalid config")
		return ikemen.MatchResult{}, nil
	}, nil); err == nil {
		t.Fatal("Run() error = nil, want invalid AI level")
	}
}
