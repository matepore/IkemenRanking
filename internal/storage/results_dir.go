package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureResultsDir creates the results directory beside the application executable.
func EnsureResultsDir() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not locate application executable: %w", err)
	}
	return EnsureResultsDirForExecutable(executablePath)
}

// EnsureResultsDirForExecutable is the path-injectable form used by tests.
func EnsureResultsDirForExecutable(executablePath string) (string, error) {
	if executablePath == "" {
		return "", fmt.Errorf("application executable path is empty")
	}
	dir := filepath.Join(filepath.Dir(executablePath), "results")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("could not create results directory: %w", err)
	}
	return dir, nil
}
