package tournament

import (
	"fmt"
	"strings"
)

// Match represents one fight between two tournament participants.
type Match struct {
	Player1 string
	Player2 string
}

// GenerateMatches creates one fight for every unique unordered participant pairing.
func GenerateMatches(participants []string) ([]Match, error) {
	if len(participants) < 2 {
		return nil, fmt.Errorf("at least two participants are required")
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

	participantCount := len(participants)
	maxInt := int(^uint(0) >> 1)
	if participantCount > maxInt/(participantCount-1) {
		return nil, fmt.Errorf("participant count is too large")
	}
	matchupCount := participantCount * (participantCount - 1) / 2
	matches := make([]Match, 0, matchupCount)
	for i := 0; i < len(participants); i++ {
		for j := i + 1; j < len(participants); j++ {
			matches = append(matches, Match{Player1: participants[i], Player2: participants[j]})
		}
	}

	return matches, nil
}
