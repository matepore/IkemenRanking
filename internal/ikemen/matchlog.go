package ikemen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	player1WinsPattern = regexp.MustCompile(`(?m)\[p1wins\]\s*=>\s*(\d+)`)
	player2WinsPattern = regexp.MustCompile(`(?m)\[p2wins\]\s*=>\s*(\d+)`)
	drawsPattern       = regexp.MustCompile(`(?m)\[("draws"|'draws'|draws)\]\s*=>\s*(\d+)`)
	winsTablePattern   = regexp.MustCompile(`(?m)\[("wins"|'wins'|wins)\]\s*=>[^\n]*\{`)
	p1ArrayIndex       = regexp.MustCompile(`(?m)\[1\]\s*=>\s*(\d+)`)
	p2ArrayIndex       = regexp.MustCompile(`(?m)\[2\]\s*=>\s*(\d+)`)
)

// ParseMatchLog extracts the aggregate win counters from an Ikemen GO text log.
//
// Legacy releases print separate p1wins/p2wins fields. Current releases print
// a wins table whose [1] and [2] entries represent P1 and P2 respectively.
// Both formats are normalized to MatchResult so the rest of the application
// does not need to know which engine version produced the log.
func ParseMatchLog(contents string) (MatchResult, error) {
	player1Wins, player2Wins, err := parseWinCounters(contents)
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

func parseWinCounters(contents string) (int, int, error) {
	// Prefer the legacy representation when both required counters are present.
	player1Wins, player1Err := parseRequiredCounter(contents, player1WinsPattern, "p1wins")
	player2Wins, player2Err := parseRequiredCounter(contents, player2WinsPattern, "p2wins")
	if player1Err == nil && player2Err == nil {
		return player1Wins, player2Wins, nil
	}

	// Newer Ikemen GO versions expose the counters as wins[1] and wins[2].
	winsTable, found := extractWinsTable(contents)
	if found {
		player1Wins, err1 := parseRequiredCounter(winsTable, p1ArrayIndex, "wins[1]")
		player2Wins, err2 := parseRequiredCounter(winsTable, p2ArrayIndex, "wins[2]")
		if err1 == nil && err2 == nil {
			return player1Wins, player2Wins, nil
		}
		if err1 != nil {
			return 0, 0, err1
		}
		return 0, 0, err2
	}

	// Preserve useful errors for legacy or unknown/incomplete log formats.
	if player1Err != nil {
		return 0, 0, player1Err
	}
	return 0, 0, player2Err
}

// extractWinsTable returns the contents of the wins table, excluding its
// outer braces. Brace counting avoids accidentally reading an unrelated [1]
// or [2] field elsewhere in the match log.
func extractWinsTable(contents string) (string, bool) {
	header := winsTablePattern.FindStringIndex(contents)
	if header == nil {
		return "", false
	}

	openBrace := strings.LastIndex(contents[header[0]:header[1]], "{") + header[0]
	depth := 0
	for i := openBrace; i < len(contents); i++ {
		switch contents[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return contents[openBrace+1 : i], true
			}
		}
	}
	return "", false
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
