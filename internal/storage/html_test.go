package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matepore/IkemenRanking/internal/ikemen"
	"github.com/matepore/IkemenRanking/internal/ranking"
	"github.com/matepore/IkemenRanking/internal/tournament"
)

func TestExportTournamentHTML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.html")
	report := TournamentReport{
		GeneratedAt: time.Date(2026, 11, 18, 12, 30, 0, 0, time.UTC),
		Installation: "C:/Games/Ikemen GO",
		Executable: "C:/Games/Ikemen GO/Ikemen_GO.exe",
		Participants: []string{"Ryu", "Ken"},
		AILevel: 8,
		RoundsToWin: 2,
		Stage: "stages/Training.def",
		Completed: 1,
		Total: 1,
		Standings: []ranking.Standing{
			{Position: 1, Character: "Ryu", Points: 1, Wins: 1, Played: 1},
			{Position: 2, Character: "Ken", Losses: 1, Played: 1},
		},
		Matches: []tournament.MatchRecord{{
			Index: 1,
			Match: tournament.Match{Player1: "Ryu", Player2: "Ken"},
			Result: ikemen.MatchResult{Player1Wins: 2, Player2Wins: 0, Winner: "Ryu", Stage: "stages/Training.def"},
		}},
	}
	if err := ExportTournamentHTML(path, report); err != nil {
		t.Fatalf("ExportTournamentHTML() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read report: %v", err)
	}
	html := string(content)
	for _, expected := range []string{
		"<!doctype html>", "Tournament Report", "Character standings",
		"Fight results", "Tournament configuration", "Ryu", "Ken",
		"2 – 0", "stages/Training.def",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("report does not contain %q", expected)
		}
	}
}

func TestExportTournamentHTMLEscapesUserContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.html")
	report := TournamentReport{
		Participants: []string{"<script>alert(1)</script>", "Ken"},
		AILevel: 8,
		RoundsToWin: 2,
		Standings: []ranking.Standing{{Position: 1, Character: "<script>alert(1)</script>"}},
	}
	if err := ExportTournamentHTML(path, report); err != nil {
		t.Fatalf("ExportTournamentHTML() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "<script>alert(1)</script>") {
		t.Fatal("report contains unescaped user content")
	}
	if !strings.Contains(string(content), "&lt;script&gt;") {
		t.Fatal("report does not HTML-escape participant names")
	}
}
