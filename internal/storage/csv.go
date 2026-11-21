package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/matepore/IkemenRanking/internal/ranking"
	"github.com/matepore/IkemenRanking/internal/tournament"
)

// ExportMatchResultsCSV writes the current saved result for each unique matchup.
func ExportMatchResultsCSV(filePath string, records []tournament.MatchRecord) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create match results CSV: %w", err)
	}
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{
		"Fight", "Player 1", "Player 2",
		"Player 1 round wins", "Player 2 round wins", "Draws",
		"Winner", "Is draw", "Stage",
	}); err != nil {
		_ = file.Close()
		return fmt.Errorf("could not write match results CSV header: %w", err)
	}
	for _, record := range records {
		winner := record.Result.Winner
		if record.Result.IsDraw {
			winner = ""
		}
		row := []string{
			strconv.Itoa(record.Index),
			record.Match.Player1,
			record.Match.Player2,
			strconv.Itoa(record.Result.Player1Wins),
			strconv.Itoa(record.Result.Player2Wins),
			strconv.Itoa(record.Result.Draws),
			winner,
			strconv.FormatBool(record.Result.IsDraw),
			record.Result.Stage,
		}
		if err := writer.Write(row); err != nil {
			_ = file.Close()
			return fmt.Errorf("could not write match results CSV row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = file.Close()
		return fmt.Errorf("could not flush match results CSV: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("could not close match results CSV: %w", err)
	}
	return nil
}

// ExportStandingsCSV writes the current global character standings.
func ExportStandingsCSV(filePath string, standings []ranking.Standing) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create standings CSV: %w", err)
	}
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{
		"Position", "Character", "Points", "Wins", "Losses", "Draws", "Played",
	}); err != nil {
		_ = file.Close()
		return fmt.Errorf("could not write standings CSV header: %w", err)
	}
	for _, standing := range standings {
		row := []string{
			strconv.Itoa(standing.Position),
			standing.Character,
			strconv.Itoa(standing.Points),
			strconv.Itoa(standing.Wins),
			strconv.Itoa(standing.Losses),
			strconv.Itoa(standing.Draws),
			strconv.Itoa(standing.Played),
		}
		if err := writer.Write(row); err != nil {
			_ = file.Close()
			return fmt.Errorf("could not write standings CSV row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = file.Close()
		return fmt.Errorf("could not flush standings CSV: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("could not close standings CSV: %w", err)
	}
	return nil
}
