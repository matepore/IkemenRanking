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


func TestParseFirstStage(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "first stage entry",
			data: "[Info]\\nname = Test\\n[Characters]\\nKFM\\n[ExtraStages]\\nstages/Training.def\\nstages/Other.def\\n",
			want: "stages/Training.def",
		},
		{
			name: "comments and blank lines",
			data: "[ExtraStages]\\n; comment\\n\\nstages/A.def ; first stage\\nstages/B.def\\n",
			want: "stages/A.def",
		},
		{
			name: "first comma-separated field",
			data: "[ExtraStages]\\nstages/A.def, displayname=Training\\nstages/B.def\\n",
			want: "stages/A.def",
		},
		{
			name: "case-insensitive section",
			data: "[extrastages]\\ncustom.def\\n",
			want: "custom.def",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseFirstStage(strings.NewReader(test.data))
			if err != nil {
				t.Fatalf("ParseFirstStage() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("ParseFirstStage() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestParseFirstStageMissingOrEmptySection(t *testing.T) {
	tests := []struct {
		name string
		data string
		want error
	}{
		{name: "missing section", data: "[Characters]\\nKFM\\n", want: ErrExtraStagesSectionNotFound},
		{name: "empty section", data: "[ExtraStages]\\n; no stages\\n[Info]\\n", want: ErrNoStagesFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseFirstStage(strings.NewReader(test.data))
			if !errors.Is(err, test.want) {
				t.Fatalf("ParseFirstStage() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestLoadFirstStage(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[Characters]\\nKFM\\n[ExtraStages]\\nstages/First.def\\nstages/Second.def\\n"
	if err := os.WriteFile(filepath.Join(dataDir, "select.def"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFirstStage(root)
	if err != nil {
		t.Fatalf("LoadFirstStage() error = %v", err)
	}
	if got != "stages/First.def" {
		t.Fatalf("LoadFirstStage() = %q, want %q", got, "stages/First.def")
	}
}
