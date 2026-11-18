package tournament

import (
	"reflect"
	"testing"
)

func TestGenerateMatches(t *testing.T) {
	participants := []string{"Ryu", "Ken", "Chun-Li"}
	got, err := GenerateMatches(participants, 2)
	if err != nil {
		t.Fatalf("GenerateMatches() error = %v", err)
	}

	want := []Match{
		{Player1: "Ryu", Player2: "Ken", Number: 1},
		{Player1: "Ryu", Player2: "Ken", Number: 2},
		{Player1: "Ryu", Player2: "Chun-Li", Number: 1},
		{Player1: "Ryu", Player2: "Chun-Li", Number: 2},
		{Player1: "Ken", Player2: "Chun-Li", Number: 1},
		{Player1: "Ken", Player2: "Chun-Li", Number: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GenerateMatches() = %#v, want %#v", got, want)
	}
}

func TestGenerateMatchesWithOneFightPerPair(t *testing.T) {
	got, err := GenerateMatches([]string{"A", "B", "C", "D"}, 1)
	if err != nil {
		t.Fatalf("GenerateMatches() error = %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("got %d fights, want 6", len(got))
	}
}

func TestGenerateMatchesRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name         string
		participants []string
		perPair      int
	}{
		{name: "no participants", participants: nil, perPair: 1},
		{name: "one participant", participants: []string{"Ryu"}, perPair: 1},
		{name: "zero fights", participants: []string{"Ryu", "Ken"}, perPair: 0},
		{name: "negative fights", participants: []string{"Ryu", "Ken"}, perPair: -1},
		{name: "empty participant", participants: []string{"Ryu", " "}, perPair: 1},
		{name: "duplicate participant", participants: []string{"Ryu", "ryu"}, perPair: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := GenerateMatches(test.participants, test.perPair); err == nil {
				t.Fatal("GenerateMatches() error = nil, want an error")
			}
		})
	}
}
