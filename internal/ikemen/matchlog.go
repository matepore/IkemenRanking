package ikemen

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	player1WinsPattern = regexp.MustCompile(`(?m)\[p1wins\]\s*=>\s*(\d+)`)
	player2WinsPattern = regexp.MustCompile(`(?m)\[p2wins\]\s*=>\s*(\d+)`)
	drawsPattern       = regexp.MustCompile(`(?m)\[draws\]\s*=>\s*(\d+)`)
)

// ParseMatchLog extracts the aggregate win counters from Ikemen GO's Lua-table
// style log. It intentionally does not treat a missing or malformed counter as
// a draw or a valid match result.
func ParseMatchLog(contents string) (MatchResult, error) {
	player1Wins, err := parseRequiredCounter(contents, player1WinsPattern, "p1wins")
	if err != nil {
		return MatchResult{}, err
	}
	player2Wins, err := parseRequiredCounter(contents, player2WinsPattern, "p2wins")
	if err != nil {
		return MatchResult{}, err
	}

	draws := 0
	if match := drawsPattern.FindStringSubmatch(contents); match != nil {
		draws, err = strconv.Atoi(match[1])
		if err != nil {
			return MatchResult{}, fmt.Errorf("invalid draws counter: %w", err)
		}
	}

	return MatchResult{
		Player1Wins: player1Wins,
		Player2Wins: player2Wins,
		Draws:       draws,
	}, nil
}

func parseRequiredCounter(contents string, pattern *regexp.Regexp, name string) (int, error) {
	match := pattern.FindStringSubmatch(contents)
	if match == nil {
		return 0, fmt.Errorf("match log does not contain a valid %s counter", name)
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("invalid %s counter: %w", name, err)
	}
	return value, nil
}
