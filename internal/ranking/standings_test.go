package ranking

import (
	"reflect"
	"testing"
)

func TestCalculateStandings(t *testing.T) {
	participants := []string{"Ryu", "Ken", "Chun-Li"}
	outcomes := []MatchOutcome{
		{Player1: "Ryu", Player2: "Ken", Winner: "Ryu"},
		{Player1: "Ryu", Player2: "Chun-Li", IsDraw: true},
		{Player1: "Ken", Player2: "Chun-Li", Winner: "Chun-Li"},
		{Player1: "Ken", Player2: "Ryu", Winner: "Ken"},
	}
	got, err := Calculate(participants, outcomes)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}
	want := []Standing{
		{Position: 1, Character: "Ryu", Points: 1, Wins: 1, Losses: 1, Draws: 1, Played: 3},
		{Position: 1, Character: "Ken", Points: 1, Wins: 1, Losses: 1, Draws: 0, Played: 2},
		{Position: 3, Character: "Chun-Li", Points: 1, Wins: 1, Losses: 0, Draws: 1, Played: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Calculate() = %#v, want %#v", got, want)
	}
}

func TestCalculateRejectsInvalidResults(t *testing.T) {
	_, err := Calculate([]string{"A", "B"}, []MatchOutcome{{Player1: "A", Player2: "B", Winner: "C"}})
	if err == nil {
		t.Fatal("Calculate() error = nil, want invalid winner")
	}
}
