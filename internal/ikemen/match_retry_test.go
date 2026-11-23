package ikemen

import (
	"errors"
	"testing"
)

func TestRunWithCrashRetriesRetriesUnexpectedExit(t *testing.T) {
	attempts := 0
	want := MatchResult{Player1Wins: 2, Player2Wins: 1}
	got, err := runWithCrashRetries(true, func() (MatchResult, error) {
		attempts++
		if attempts < 3 {
			return MatchResult{}, &processExitError{details: "simulated crash", logPath: t.TempDir() + "/failed.log"}
		}
		return want, nil
	})
	if err != nil {
		t.Fatalf("runWithCrashRetries() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if got != want {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
}

func TestRunWithCrashRetriesCanBeDisabled(t *testing.T) {
	attempts := 0
	_, err := runWithCrashRetries(false, func() (MatchResult, error) {
		attempts++
		return MatchResult{}, &processExitError{details: "simulated crash"}
	})
	if err == nil {
		t.Fatal("runWithCrashRetries() error = nil, want error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestRunWithCrashRetriesDoesNotRetryNonProcessErrors(t *testing.T) {
	attempts := 0
	wantErr := errors.New("match log parse failure")
	_, err := runWithCrashRetries(true, func() (MatchResult, error) {
		attempts++
		return MatchResult{}, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestRunWithCrashRetriesStopsAfterMaximumAttempts(t *testing.T) {
	attempts := 0
	_, err := runWithCrashRetries(true, func() (MatchResult, error) {
		attempts++
		return MatchResult{}, &processExitError{details: "simulated crash"}
	})
	if err == nil {
		t.Fatal("runWithCrashRetries() error = nil, want error")
	}
	if want := MaxCrashRetries + 1; attempts != want {
		t.Fatalf("attempts = %d, want %d", attempts, want)
	}
}
