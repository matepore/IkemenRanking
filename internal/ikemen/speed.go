package ikemen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GameSpeedMode identifies the command-line speed convention used by Ikemen GO.
type GameSpeedMode int

const (
	// GameSpeedModeSteps is used by Ikemen GO v1.0.0 and newer.
	GameSpeedModeSteps GameSpeedMode = iota
	// GameSpeedModePercentage is used by legacy releases such as v0.99.0.
	GameSpeedModePercentage
)

func (mode GameSpeedMode) String() string {
	switch mode {
	case GameSpeedModeSteps:
		return "Step-based (v1.0.0+)"
	case GameSpeedModePercentage:
		return "Percentage-based (legacy)"
	default:
		return "Unknown"
	}
}

// DetectGameSpeedMode inspects main.lua to detect the installed engine's actual
// -speed convention. This also supports custom builds whose executable metadata
// may not identify their source version.
func DetectGameSpeedMode(installationDir string) (GameSpeedMode, error) {
	if strings.TrimSpace(installationDir) == "" {
		return 0, fmt.Errorf("Ikemen GO installation directory is required")
	}
	scriptPath := filepath.Join(installationDir, "external", "script", "main.lua")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return 0, fmt.Errorf("could not read Ikemen GO speed configuration at %q: %w", scriptPath, err)
	}
	return ParseGameSpeedMode(string(content))
}

// ParseGameSpeedMode identifies the -speed convention from the engine Lua script.
func ParseGameSpeedMode(source string) (GameSpeedMode, error) {
	source = strings.ToLower(source)
	legacy := regexp.MustCompile(`(?s)setgamespeed\s*\(\s*tonumber\s*\([^)]*-speed[^)]*\)\s*\*\s*config\.framerate\s*/\s*100\s*\)`)
	if legacy.MatchString(source) || (strings.Contains(source, "config.framerate / 100") && strings.Contains(source, "setgamespeed")) {
		return GameSpeedModePercentage, nil
	}
	modern := regexp.MustCompile(`(?s)speed_input\s*>=\s*-9\s+and\s+speed_input\s*<=\s*9`)
	if modern.MatchString(source) && strings.Contains(source, "target_game_speed") && strings.Contains(source, "setgamespeed(target_game_speed)") {
		return GameSpeedModeSteps, nil
	}
	return 0, fmt.Errorf("unrecognized Ikemen GO -speed implementation in external/script/main.lua")
}

// ValidateGameSpeed validates a selected value for the detected engine mode.
func ValidateGameSpeed(mode GameSpeedMode, value int) error {
	switch mode {
	case GameSpeedModeSteps:
		if value < 0 || value > 9 {
			return fmt.Errorf("game speed for this Ikemen GO version must be between 0 and 9")
		}
	case GameSpeedModePercentage:
		if value < 100 || value > 200 {
			return fmt.Errorf("game speed for this legacy Ikemen GO version must be between 100%% and 200%%")
		}
	default:
		return fmt.Errorf("Ikemen GO game speed mode has not been detected")
	}
	return nil
}

// GameSpeedOptions returns UI labels and CLI values for the detected engine mode.
func GameSpeedOptions(mode GameSpeedMode) ([]string, map[string]int, string, error) {
	options := make([]string, 0, 11)
	values := make(map[string]int, 11)
	switch mode {
	case GameSpeedModeSteps:
		options = append(options, "Normal")
		values["Normal"] = 0
		for value := 1; value <= 9; value++ {
			label := fmt.Sprintf("Faster (+%d)", value)
			options = append(options, label)
			values[label] = value
		}
		return options, values, "Normal", nil
	case GameSpeedModePercentage:
		for value := 100; value <= 200; value += 10 {
			label := fmt.Sprintf("Faster (%d%%)", value)
			if value == 100 {
				label = "Normal (100%)"
			}
			options = append(options, label)
			values[label] = value
		}
		return options, values, "Normal (100%)", nil
	default:
		return nil, nil, "", fmt.Errorf("cannot create game speed options for an unknown Ikemen GO version")
	}
}
