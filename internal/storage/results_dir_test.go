package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureResultsDirForExecutable(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "bin", "ikemen-ranking.exe")
	want := filepath.Join(filepath.Dir(executable), "results")
	got, err := EnsureResultsDirForExecutable(executable)
	if err != nil {
		t.Fatalf("EnsureResultsDirForExecutable() error = %v", err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("results directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", got)
	}
}

func TestEnsureResultsDirForExecutableRejectsEmptyPath(t *testing.T) {
	if _, err := EnsureResultsDirForExecutable(""); err == nil {
		t.Fatal("expected an error for an empty executable path")
	}
}
