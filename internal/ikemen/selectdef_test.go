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


func TestParseStages(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "all stage entries in order",
			data: "[Info]\nname = Test\n[ExtraStages]\nstages/Training.def\nstages/Other.def\n[Characters]\nKFM\n",
			want: []string{"stages/Training.def", "stages/Other.def"},
		},
		{
			name: "comments blank lines and duplicate paths",
			data: "[ExtraStages]\n; comment\n\nstages/A.def ; first\nstages/B.def, displayname=B\nSTAGES/a.DEF\n",
			want: []string{"stages/A.def", "stages/B.def"},
		},
		{
			name: "case-insensitive section",
			data: "[extrastages]\ncustom.def\n",
			want: []string{"custom.def"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseStages(strings.NewReader(test.data))
			if err != nil {
				t.Fatalf("ParseStages() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseStages() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseStagesMissingOrEmptySection(t *testing.T) {
	tests := []struct {
		name string
		data string
		want error
	}{
		{name: "missing section", data: "[Characters]\nKFM\n", want: ErrExtraStagesSectionNotFound},
		{name: "empty section", data: "[ExtraStages]\n; no stages\n[Info]\n", want: ErrNoStagesFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseStages(strings.NewReader(test.data))
			if !errors.Is(err, test.want) {
				t.Fatalf("ParseStages() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestLoadStages(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[Characters]\nKFM\n[ExtraStages]\nstages/First.def\nstages/Second.def\n"
	if err := os.WriteFile(filepath.Join(dataDir, "select.def"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadStages(root)
	if err != nil {
		t.Fatalf("LoadStages() error = %v", err)
	}
	want := []string{"stages/First.def", "stages/Second.def"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadStages() = %#v, want %#v", got, want)
	}
}
