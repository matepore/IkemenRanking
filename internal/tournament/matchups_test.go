package tournament

import (
	"reflect"
	"testing"
)

func TestGenerateMatches(t *testing.T) {
	participants := []string{"Ryu", "Ken", "Chun-Li"}
	got, err := GenerateMatches(participants)
	if err != nil {
		t.Fatalf("GenerateMatches() error = %v", err)
	}
	want := []Match{
		{Player1: "Ryu", Player2: "Ken"},
		{Player1: "Ryu", Player2: "Chun-Li"},
		{Player1: "Ken", Player2: "Chun-Li"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GenerateMatches() = %#v, want %#v", got, want)
	}
}

func TestGenerateMatchesCreatesOneFightPerPair(t *testing.T) {
	got, err := GenerateMatches([]string{"A", "B", "C", "D"})
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
	}{
		{name: "no participants", participants: nil},
		{name: "one participant", participants: []string{"Ryu"}},
		{name: "empty participant", participants: []string{"Ryu", " "}},
		{name: "duplicate participant", participants: []string{"Ryu", "ryu"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := GenerateMatches(test.participants); err == nil {
				t.Fatal("GenerateMatches() error = nil, want an error")
			}
		})
	}
}
