package tournament

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matepore/IkemenRanking/internal/ikemen"
)

func awaitSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func TestRunPersistentWithControlPausesBetweenFightsAndResumes(t *testing.T) {
	store := newMemoryFightStore()
	controller := &PauseController{}
	firstStarted := make(chan struct{})
	finishFirst := make(chan struct{})
	paused := make(chan struct{})
	done := make(chan error, 1)
	var calls atomic.Int32

	go func() {
		_, err := RunPersistentWithControl(
			context.Background(), persistentTestConfig(), store, false,
			func(_ context.Context, match ikemen.MatchConfig) (ikemen.MatchResult, error) {
				if calls.Add(1) == 1 {
					close(firstStarted)
					<-finishFirst
				}
				return ikemen.MatchResult{Winner: match.Player1}, nil
			}, nil, controller, func() { close(paused) },
		)
		done <- err
	}()

	awaitSignal(t, firstStarted, "first fight to start")
	controller.Pause()
	close(finishFirst)
	awaitSignal(t, paused, "tournament to pause after first fight")

	if got := calls.Load(); got != 1 {
		t.Fatalf("runner started %d fights before pause, want exactly 1", got)
	}
	if len(store.records) != 1 {
		t.Fatalf("persisted %d fights at pause, want 1", len(store.records))
	}

	controller.Resume()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunPersistentWithControl() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("tournament did not finish after resume")
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("runner executed %d fights, want 3", got)
	}
}

func TestRunPersistentWithControlCancellationWhilePaused(t *testing.T) {
	store := newMemoryFightStore()
	controller := &PauseController{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	paused := make(chan struct{})
	done := make(chan error, 1)

	controller.Pause()
	go func() {
		_, err := RunPersistentWithControl(
			ctx, persistentTestConfig(), store, false,
			func(_ context.Context, match ikemen.MatchConfig) (ikemen.MatchResult, error) {
				return ikemen.MatchResult{Winner: match.Player1}, nil
			}, nil, controller, func() { close(paused) },
		)
		done <- err
	}()

	awaitSignal(t, paused, "runner to enter paused state")
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("RunPersistentWithControl() error = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runner did not stop after cancellation")
	}
	if len(store.records) != 0 {
		t.Fatalf("persisted %d fights, want none because cancellation occurred before the first fight", len(store.records))
	}
}
