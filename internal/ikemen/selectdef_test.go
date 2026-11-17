package ikemen

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseSelectDef(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "normal entries and extra parameters",
			data: "[Info]\nname = Test\n[Characters]\nccijhun, stages/Jhun.def, order=1\nKrohnen, order=2\n[ExtraStages]\nstage.def\n",
			want: []string{"ccijhun", "Krohnen"},
		},
		{
			name: "comments blank lines random select and duplicates",
			data: "\uFEFF[Characters]\n; full line comment\n\nrandomselect\nrandomselect, order=1\nJhun ; inline comment\njhun, order=3\nRalf, stages/Ralf.def\n",
			want: []string{"Jhun", "Ralf"},
		},
		{
			name: "section name is case insensitive",
			data: "[characters]\nKFM\n",
			want: []string{"KFM"},
		},
		{
			name: "empty characters section",
			data: "[Characters]\n[ExtraStages]\n",
			want: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseSelectDef(strings.NewReader(test.data))
			if err != nil {
				t.Fatalf("ParseSelectDef() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseSelectDef() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseSelectDefMissingCharactersSection(t *testing.T) {
	_, err := ParseSelectDef(strings.NewReader("[Info]\nname = Test\n"))
	if !errors.Is(err, ErrCharactersSectionNotFound) {
		t.Fatalf("expected ErrCharactersSectionNotFound, got %v", err)
	}
}

func TestLoadCharacters(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[Characters]\nccijhun\nKrohnen\n"
	if err := os.WriteFile(filepath.Join(dataDir, "select.def"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadCharacters(root)
	if err != nil {
		t.Fatalf("LoadCharacters() error = %v", err)
	}
	want := []string{"ccijhun", "Krohnen"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadCharacters() = %#v, want %#v", got, want)
	}
}

func TestLoadCharactersMissingFile(t *testing.T) {
	_, err := LoadCharacters(t.TempDir())
	if err == nil {
		t.Fatal("expected an error for a missing select.def")
	}
}
