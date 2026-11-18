package ikemen

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrCharactersSectionNotFound = errors.New("the [Characters] section was not found in data/select.def")
var ErrExtraStagesSectionNotFound = errors.New("the [ExtraStages] section was not found in data/select.def")
var ErrNoStagesFound = errors.New("the [ExtraStages] section does not contain any stage entries")

const RandomStage = "Random"

// LoadCharacters reads the character identifiers from an Ikemen GO installation.
func LoadCharacters(installationDir string) ([]string, error) {
	selectDefPath := filepath.Join(installationDir, "data", "select.def")

	info, err := os.Stat(selectDefPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("could not find %s: %w", selectDefPath, err)
		}
		return nil, fmt.Errorf("could not access %s: %w", selectDefPath, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", selectDefPath)
	}

	file, err := os.Open(selectDefPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", selectDefPath, err)
	}
	defer file.Close()

	characters, err := ParseSelectDef(file)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", selectDefPath, err)
	}
	return characters, nil
}

// ParseSelectDef extracts character identifiers from the [Characters] section.
// It preserves the identifier as written in the file and ignores comments,
// blank lines, random-select entries, and case-insensitive duplicates.
func ParseSelectDef(source io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	inCharacters := false
	foundCharacters := false
	seen := make(map[string]struct{})
	characters := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			if inCharacters {
				break
			}
			if strings.EqualFold(section, "Characters") {
				inCharacters = true
				foundCharacters = true
			}
			continue
		}

		if !inCharacters {
			continue
		}

		if comment := strings.IndexByte(line, ';'); comment >= 0 {
			line = strings.TrimSpace(line[:comment])
		}
		if line == "" {
			continue
		}

		// In select.def, the first comma-separated field is the character ID.
		id := strings.TrimSpace(strings.SplitN(line, ",", 2)[0])
		if id == "" || strings.EqualFold(id, "randomselect") {
			continue
		}

		key := strings.ToLower(id)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		characters = append(characters, id)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading select.def: %w", err)
	}
	if !foundCharacters {
		return nil, ErrCharactersSectionNotFound
	}
	return characters, nil
}


// LoadStages reads every stage identifier from the [ExtraStages] section.
func LoadStages(installationDir string) ([]string, error) {
	selectDefPath := filepath.Join(installationDir, "data", "select.def")
	info, err := os.Stat(selectDefPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("could not find %s: %w", selectDefPath, err)
		}
		return nil, fmt.Errorf("could not access %s: %w", selectDefPath, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", selectDefPath)
	}

	file, err := os.Open(selectDefPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", selectDefPath, err)
	}
	defer file.Close()

	stages, err := ParseStages(file)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", selectDefPath, err)
	}
	return stages, nil
}

// LoadFirstStage is kept as a convenience wrapper for callers that need the
// first configured stage.
func LoadFirstStage(installationDir string) (string, error) {
	stages, err := LoadStages(installationDir)
	if err != nil {
		return "", err
	}
	return stages[0], nil
}

// ParseStages returns all non-empty stage entries in [ExtraStages], preserving
// their order and spelling. Comments, blank lines, and case-insensitive
// duplicate paths are ignored.
func ParseStages(source io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	inStages := false
	foundStages := false
	seen := make(map[string]struct{})
	stages := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\\uFEFF"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if inStages {
				break
			}
			if strings.EqualFold(strings.TrimSpace(line[1:len(line)-1]), "ExtraStages") {
				inStages = true
				foundStages = true
			}
			continue
		}
		if !inStages {
			continue
		}
		if comment := strings.IndexByte(line, ';'); comment >= 0 {
			line = strings.TrimSpace(line[:comment])
		}
		if line == "" {
			continue
		}
		stage := strings.TrimSpace(strings.SplitN(line, ",", 2)[0])
		if stage == "" {
			continue
		}
		key := strings.ToLower(stage)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		stages = append(stages, stage)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading select.def: %w", err)
	}
	if !foundStages {
		return nil, ErrExtraStagesSectionNotFound
	}
	if len(stages) == 0 {
		return nil, ErrNoStagesFound
	}
	return stages, nil
}
