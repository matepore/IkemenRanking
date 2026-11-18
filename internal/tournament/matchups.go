package tournament

import (
	"fmt"
	"strings"
)

// Match represents one individual fight between two tournament participants.
// Number is the fight's one-based position within the matchup.
type Match struct {
	Player1 string
	Player2 string
	Number  int
}

// GenerateMatches creates every unique unordered participant pairing and
// repeats each pairing matchesPerPair times.
func GenerateMatches(participants []string, matchesPerPair int) ([]Match, error) {
	if len(participants) < 2 {
		return nil, fmt.Errorf("at least two participants are required")
	}
	if matchesPerPair < 1 {
		return nil, fmt.Errorf("matches per matchup must be at least 1")
	}

	seen := make(map[string]struct{}, len(participants))
	for _, participant := range participants {
		name := strings.TrimSpace(participant)
		if name == "" {
			return nil, fmt.Errorf("participant names cannot be empty")
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate participant: %s", participant)
		}
		seen[key] = struct{}{}
	}

	matchupCount := len(participants) * (len(participants) - 1) / 2
	matches := make([]Match, 0, matchupCount*matchesPerPair)

	for i := 0; i < len(participants); i++ {
		for j := i + 1; j < len(participants); j++ {
			for number := 1; number <= matchesPerPair; number++ {
				matches = append(matches, Match{
					Player1: participants[i],
					Player2: participants[j],
					Number:  number,
				})
			}
		}
	}

	return matches, nil
}
