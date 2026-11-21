package ranking

import (
	"fmt"
	"sort"
	"strings"
)

// MatchOutcome is the minimal information needed to calculate standings.
type MatchOutcome struct {
	Player1 string
	Player2 string
	Winner  string
	IsDraw  bool
}

// Standing contains the accumulated tournament statistics for one character.
type Standing struct {
	Position  int
	Character string
	Points    int
	Wins      int
	Losses    int
	Draws     int
	Played    int
}

// Calculate builds a points-based ranking. A win awards one point; a draw
// awards no points to either participant. Equal point totals share a position.
func Calculate(participants []string, outcomes []MatchOutcome) ([]Standing, error) {
	standings := make(map[string]*Standing, len(participants))
	order := make([]string, 0, len(participants))
	for _, participant := range participants {
		name := strings.TrimSpace(participant)
		if name == "" {
			return nil, fmt.Errorf("participant names cannot be empty")
		}
		key := strings.ToLower(name)
		if _, exists := standings[key]; exists {
			return nil, fmt.Errorf("duplicate participant: %s", participant)
		}
		standings[key] = &Standing{Character: participant}
		order = append(order, key)
	}

	for index, outcome := range outcomes {
		p1Key, p2Key := strings.ToLower(outcome.Player1), strings.ToLower(outcome.Player2)
		p1, p1OK := standings[p1Key]
		p2, p2OK := standings[p2Key]
		if !p1OK || !p2OK {
			return nil, fmt.Errorf("result %d contains an unknown participant", index+1)
		}
		if p1Key == p2Key {
			return nil, fmt.Errorf("result %d contains the same participant twice", index+1)
		}
		p1.Played++
		p2.Played++
		if outcome.IsDraw {
			if outcome.Winner != "" {
				return nil, fmt.Errorf("result %d is a draw but also declares a winner", index+1)
			}
			p1.Draws++
			p2.Draws++
			continue
		}

		winnerKey := strings.ToLower(outcome.Winner)
		switch winnerKey {
		case p1Key:
			p1.Wins++
			p1.Points++
			p2.Losses++
		case p2Key:
			p2.Wins++
			p2.Points++
			p1.Losses++
		default:
			return nil, fmt.Errorf("result %d has no valid winner", index+1)
		}
	}

	result := make([]Standing, 0, len(order))
	for _, key := range order {
		result = append(result, *standings[key])
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Points > result[j].Points
	})
	position := 0
	for i := range result {
		if i == 0 || result[i].Points != result[i-1].Points {
			position = i + 1
		}
		result[i].Position = position
	}
	return result, nil
}
