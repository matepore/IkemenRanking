package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/matepore/IkemenRanking/internal/ikemen"
	"github.com/matepore/IkemenRanking/internal/tournament"
)

func TestDatabasePersistsAndNormalizesFight(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "results", "ikemen-ranking.db")

	database, err := OpenDatabase(path)
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	record := tournament.MatchRecord{
		Match: tournament.Match{Player1: "B", Player2: "A"},
		Result: ikemen.MatchResult{
			Player1Wins: 1, Player2Wins: 2, Draws: 1,
			Winner: "A", Stage: "Training Room",
		},
	}
	if err := database.UpsertFight(ctx, record); err != nil {
		t.Fatalf("UpsertFight() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	database, err = OpenDatabase(path)
	if err != nil {
		t.Fatalf("reopening database: %v", err)
	}
	defer database.Close()

	records, err := database.ListFightsForParticipants(ctx, []string{"A", "B"})
	if err != nil {
		t.Fatalf("ListFightsForParticipants() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	got := records[0]
	if got.Match.Player1 != "A" || got.Match.Player2 != "B" {
		t.Fatalf("stored matchup = %q vs %q, want A vs B", got.Match.Player1, got.Match.Player2)
	}
	if got.Result.Player1Wins != 2 || got.Result.Player2Wins != 1 {
		t.Fatalf("stored score = %d-%d, want 2-1", got.Result.Player1Wins, got.Result.Player2Wins)
	}
	if got.Result.Draws != 1 || got.Result.Winner != "A" || got.Result.Stage != "Training Room" {
		t.Fatalf("stored result fields were not preserved: %#v", got.Result)
	}
}

func TestDatabaseUpsertReplacesSameUnorderedPair(t *testing.T) {
	ctx := context.Background()
	database, err := OpenDatabase(filepath.Join(t.TempDir(), "ikemen-ranking.db"))
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	defer database.Close()

	first := tournament.MatchRecord{
		Match: tournament.Match{Player1: "A", Player2: "B"},
		Result: ikemen.MatchResult{Player1Wins: 2, Winner: "A", Stage: "Stage 1"},
	}
	second := tournament.MatchRecord{
		Match: tournament.Match{Player1: "B", Player2: "A"},
		Result: ikemen.MatchResult{Player2Wins: 2, Winner: "A", Stage: "Stage 2"},
	}
	if err := database.UpsertFight(ctx, first); err != nil {
		t.Fatalf("first UpsertFight() error = %v", err)
	}
	if err := database.UpsertFight(ctx, second); err != nil {
		t.Fatalf("second UpsertFight() error = %v", err)
	}

	records, err := database.ListFightsForParticipants(ctx, []string{"A", "B"})
	if err != nil {
		t.Fatalf("ListFightsForParticipants() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want exactly 1 for the unordered pair", len(records))
	}
	if records[0].Result.Player1Wins != 2 || records[0].Result.Player2Wins != 0 {
		t.Fatalf("updated score = %d-%d, want 2-0", records[0].Result.Player1Wins, records[0].Result.Player2Wins)
	}
	if records[0].Result.Stage != "Stage 2" {
		t.Fatalf("updated stage = %q, want Stage 2", records[0].Result.Stage)
	}
}

func TestDatabaseListsOnlyFightsBetweenSelectedParticipants(t *testing.T) {
	ctx := context.Background()
	database, err := OpenDatabase(filepath.Join(t.TempDir(), "ikemen-ranking.db"))
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	defer database.Close()

	seed := []tournament.MatchRecord{
		{Match: tournament.Match{Player1: "A", Player2: "B"}, Result: ikemen.MatchResult{Winner: "A"}},
		{Match: tournament.Match{Player1: "A", Player2: "C"}, Result: ikemen.MatchResult{Winner: "C"}},
		{Match: tournament.Match{Player1: "B", Player2: "C"}, Result: ikemen.MatchResult{IsDraw: true}},
	}
	for _, record := range seed {
		if err := database.UpsertFight(ctx, record); err != nil {
			t.Fatalf("UpsertFight(%s vs %s) error = %v", record.Match.Player1, record.Match.Player2, err)
		}
	}

	records, err := database.ListFightsForParticipants(ctx, []string{"A", "B"})
	if err != nil {
		t.Fatalf("ListFightsForParticipants() error = %v", err)
	}
	if len(records) != 1 || records[0].Match.Player1 != "A" || records[0].Match.Player2 != "B" {
		t.Fatalf("selected participants returned unexpected records: %#v", records)
	}
}

func TestDatabaseRejectsInvalidFightAndParticipantList(t *testing.T) {
	ctx := context.Background()
	database, err := OpenDatabase(filepath.Join(t.TempDir(), "ikemen-ranking.db"))
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	defer database.Close()

	tests := []struct {
		name string
		record tournament.MatchRecord
	}{
		{"same participant", tournament.MatchRecord{Match: tournament.Match{Player1: "A", Player2: "a"}, Result: ikemen.MatchResult{Winner: "A"}}},
		{"negative counter", tournament.MatchRecord{Match: tournament.Match{Player1: "A", Player2: "B"}, Result: ikemen.MatchResult{Player1Wins: -1, Winner: "A"}}},
		{"winner is not a participant", tournament.MatchRecord{Match: tournament.Match{Player1: "A", Player2: "B"}, Result: ikemen.MatchResult{Winner: "C"}}},
		{"draw declares winner", tournament.MatchRecord{Match: tournament.Match{Player1: "A", Player2: "B"}, Result: ikemen.MatchResult{IsDraw: true, Winner: "A"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := database.UpsertFight(ctx, test.record); err == nil {
				t.Fatal("UpsertFight() error = nil, want validation error")
			}
		})
	}
	if _, err := database.ListFightsForParticipants(ctx, []string{"A", "a"}); err == nil {
		t.Fatal("ListFightsForParticipants() error = nil, want duplicate participant error")
	}
}

func TestDatabaseListsAllSavedFightsAcrossParticipantSelections(t *testing.T) {
 ctx:=context.Background()
 database,err:=OpenDatabase(filepath.Join(t.TempDir(),"ikemen-ranking.db"));if err!=nil{t.Fatal(err)};defer database.Close()
 seed:=[]tournament.MatchRecord{
  {Match:tournament.Match{Player1:"A",Player2:"B"},Result:ikemen.MatchResult{Winner:"A",Player1Wins:2}},
  {Match:tournament.Match{Player1:"C",Player2:"D"},Result:ikemen.MatchResult{Winner:"D",Player2Wins:2}},
 }
 for _,r:=range seed{if err:=database.UpsertFight(ctx,r);err!=nil{t.Fatal(err)}}
 records,err:=database.ListAllFights(ctx);if err!=nil{t.Fatal(err)}
 if len(records)!=2||records[0].Match.Player1!="A"||records[1].Match.Player1!="C"{t.Fatalf("unexpected all-fights result: %#v",records)}
 if records[0].Index!=1||records[1].Index!=2{t.Fatalf("unexpected indexes: %#v",records)}
}
func TestDatabaseReplacingWinnerKeepsOneUniqueMatchup(t *testing.T){
 ctx:=context.Background();database,err:=OpenDatabase(filepath.Join(t.TempDir(),"ikemen-ranking.db"));if err!=nil{t.Fatal(err)};defer database.Close()
 first:=tournament.MatchRecord{Match:tournament.Match{Player1:"A",Player2:"B"},Result:ikemen.MatchResult{Player1Wins:2,Winner:"A"}}
 replay:=tournament.MatchRecord{Match:tournament.Match{Player1:"B",Player2:"A"},Result:ikemen.MatchResult{Player2Wins:2,Winner:"B"}}
 if err:=database.UpsertFight(ctx,first);err!=nil{t.Fatal(err)};if err:=database.UpsertFight(ctx,replay);err!=nil{t.Fatal(err)}
 records,err:=database.ListAllFights(ctx);if err!=nil{t.Fatal(err)}
 if len(records)!=1||records[0].Result.Winner!="B"{t.Fatalf("expected one matchup with B as current winner, got %#v",records)}
}
