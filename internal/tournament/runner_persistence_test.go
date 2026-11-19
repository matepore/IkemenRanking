package tournament

import (
	"context"
	"errors"
	"testing"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

type memoryFightStore struct {
	records map[string]MatchRecord
}

func newMemoryFightStore(records ...MatchRecord) *memoryFightStore {
	store := &memoryFightStore{records: make(map[string]MatchRecord)}
	for _, record := range records {
		key, _ := matchupKey(record.Match.Player1, record.Match.Player2)
		store.records[key] = record
	}
	return store
}

func (store *memoryFightStore) ListFightsForParticipants(_ context.Context, participants []string) ([]MatchRecord, error) {
	selected := make(map[string]bool, len(participants))
	for _, participant := range participants {
		selected[participant] = true
	}
	records := make([]MatchRecord, 0, len(store.records))
	for _, record := range store.records {
		if selected[record.Match.Player1] && selected[record.Match.Player2] {
			records = append(records, record)
		}
	}
	return records, nil
}

func (store *memoryFightStore) UpsertFight(_ context.Context, record MatchRecord) error {
	key, err := matchupKey(record.Match.Player1, record.Match.Player2)
	if err != nil {
		return err
	}
	store.records[key] = record
	return nil
}

func persistentTestConfig() Config {
	return Config{Participants: []string{"A", "B", "C"}, AILevel: 8, RoundsToWin: 2, Stage: "Random"}
}

func TestRunPersistentSkipsPreviouslyCompletedFights(t *testing.T) {
	store := newMemoryFightStore(MatchRecord{
		Match: Match{Player1: "A", Player2: "B"},
		Result: ikemen.MatchResult{Player1Wins: 2, Winner: "A"},
	})
	calls := 0
	progress := make([]Progress, 0)
	records, err := RunPersistent(context.Background(), persistentTestConfig(), store, false,
		func(_ context.Context, match ikemen.MatchConfig) (ikemen.MatchResult, error) {
			calls++
			return ikemen.MatchResult{Winner: match.Player1, Player1Wins: 2}, nil
		},
		func(update Progress) { progress = append(progress, update) },
	)
	if err != nil {
		t.Fatalf("RunPersistent() error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("runner executed %d fights, want 2 (one saved fight should be skipped)", calls)
	}
	if len(records) != 3 {
		t.Fatalf("got %d returned records, want all 3 matchups", len(records))
	}
	if len(progress) != 3 || progress[2].Completed != 3 || progress[2].Played != 2 || progress[2].Skipped != 1 {
		t.Fatalf("unexpected final progress: %#v", progress)
	}
}

func TestRunPersistentReplayReplacesSavedResult(t *testing.T) {
	store := newMemoryFightStore(MatchRecord{
		Match: Match{Player1: "A", Player2: "B"},
		Result: ikemen.MatchResult{Player1Wins: 2, Winner: "A", Stage: "Old stage"},
	})
	calls := 0
	_, err := RunPersistent(context.Background(), persistentTestConfig(), store, true,
		func(_ context.Context, match ikemen.MatchConfig) (ikemen.MatchResult, error) {
			calls++
			return ikemen.MatchResult{Player1Wins: 2, Winner: match.Player1, Stage: "New stage"}, nil
		}, nil)
	if err != nil {
		t.Fatalf("RunPersistent() error = %v", err)
	}
	if calls != 3 {
		t.Fatalf("runner executed %d fights, want all 3", calls)
	}
	key, _ := matchupKey("A", "B")
	if got := store.records[key].Result.Stage; got != "New stage" {
		t.Fatalf("saved stage = %q, want replacement result", got)
	}
}

func TestRunPersistentWithNewParticipantRunsOnlyNewPairs(t *testing.T) {
	store := newMemoryFightStore(MatchRecord{
		Match: Match{Player1: "A", Player2: "B"},
		Result: ikemen.MatchResult{Player1Wins: 2, Winner: "A"},
	})
	config := persistentTestConfig()
	config.Participants = []string{"A", "B", "C", "D"}
	calls := 0
	records, err := RunPersistent(context.Background(), config, store, false,
		func(_ context.Context, match ikemen.MatchConfig) (ikemen.MatchResult, error) {
			calls++
			return ikemen.MatchResult{Winner: match.Player1}, nil
		}, nil)
	if err != nil {
		t.Fatalf("RunPersistent() error = %v", err)
	}
	if calls != 5 {
		t.Fatalf("runner executed %d fights, want 5 new pairs", calls)
	}
	if len(records) != 6 {
		t.Fatalf("got %d records, want all 6 pairs", len(records))
	}
}

func TestRunPersistentKeepsOldResultWhenReplayFails(t *testing.T) {
	old := MatchRecord{
		Match: Match{Player1: "A", Player2: "B"},
		Result: ikemen.MatchResult{Player1Wins: 2, Winner: "A", Stage: "Old stage"},
	}
	store := newMemoryFightStore(old)
	wantErr := errors.New("simulated game failure")
	_, err := RunPersistent(context.Background(), persistentTestConfig(), store, true,
		func(context.Context, ikemen.MatchConfig) (ikemen.MatchResult, error) {
			return ikemen.MatchResult{}, wantErr
		}, nil)
	if err == nil {
		t.Fatal("RunPersistent() error = nil, want match failure")
	}
	key, _ := matchupKey("A", "B")
	if got := store.records[key].Result.Stage; got != "Old stage" {
		t.Fatalf("failed replay replaced saved stage with %q", got)
	}
}
