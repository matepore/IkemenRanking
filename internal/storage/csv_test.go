package storage

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matepore/IkemenRanking/internal/ikemen"
	"github.com/matepore/IkemenRanking/internal/ranking"
	"github.com/matepore/IkemenRanking/internal/tournament"
)

func TestExportMatchResultsCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "matches.csv")
	records := []tournament.MatchRecord{{
		Index: 1,
		Match: tournament.Match{Player1: "A", Player2: "B"},
		Result: ikemen.MatchResult{Player1Wins: 2, Player2Wins: 1, Winner: "A", Stage: "stages/test.def"},
	}}
	if err := ExportMatchResultsCSV(path, records); err != nil {
		t.Fatalf("ExportMatchResultsCSV() error = %v", err)
	}
	file, err := os.Open(path)
	if err != nil { t.Fatal(err) }
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil { t.Fatal(err) }
	if len(rows) != 2 || rows[1][6] != "A" || rows[1][8] != "stages/test.def" {
		t.Fatalf("unexpected CSV rows: %#v", rows)
	}
}

func TestExportStandingsCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ranking.csv")
	want := []ranking.Standing{{Position: 1, Character: "A", Points: 1, Wins: 1, Played: 1}}
	if err := ExportStandingsCSV(path, want); err != nil {
		t.Fatalf("ExportStandingsCSV() error = %v", err)
	}
	file, err := os.Open(path)
	if err != nil { t.Fatal(err) }
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil { t.Fatal(err) }
	if len(rows) != 2 || rows[1][1] != "A" || rows[1][2] != "1" {
		t.Fatalf("unexpected CSV rows: %#v", rows)
	}
}

func TestExportMatchResultsCSVOverwritesPreviousExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "matches.csv")
	first := []tournament.MatchRecord{{Index: 1, Match: tournament.Match{Player1: "OldA", Player2: "OldB"}, Result: ikemen.MatchResult{Winner: "OldA"}}}
	second := []tournament.MatchRecord{{Index: 1, Match: tournament.Match{Player1: "NewA", Player2: "NewB"}, Result: ikemen.MatchResult{Winner: "NewA"}}}
	if err := ExportMatchResultsCSV(path, first); err != nil { t.Fatal(err) }
	if err := ExportMatchResultsCSV(path, second); err != nil { t.Fatal(err) }
	content, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	if strings.Contains(string(content), "OldA") || !strings.Contains(string(content), "NewA") {
		t.Fatal("CSV was not replaced with the latest export")
	}
}
