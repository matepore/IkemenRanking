package ikemen

import "testing"

func TestParseMatchLog(t *testing.T) {
	contents := `{
		[winTeam] => 1,
		[p1wins] => 0,
		[p2wins] => 2,
		[draws] => 0,
		[lastRound] => 2,
		[match] => {}
	}`

	got, err := ParseMatchLog(contents)
	if err != nil {
		t.Fatalf("ParseMatchLog() error = %v", err)
	}
	if got.Player1Wins != 0 || got.Player2Wins != 2 || got.Draws != 0 {
		t.Fatalf("unexpected counters: %#v", got)
	}
}

func TestParseMatchLogAllowsMissingDrawCounter(t *testing.T) {
	got, err := ParseMatchLog("[p1wins] => 1\n[p2wins] => 1")
	if err != nil {
		t.Fatalf("ParseMatchLog() error = %v", err)
	}
	if got.Player1Wins != 1 || got.Player2Wins != 1 || got.Draws != 0 {
		t.Fatalf("unexpected counters: %#v", got)
	}
}

func TestParseMatchLogRejectsMissingOrMalformedCounters(t *testing.T) {
	tests := []struct {
		name string
		log  string
	}{
		{name: "missing p1wins", log: "[p2wins] => 2"},
		{name: "missing p2wins", log: "[p1wins] => 2"},
		{name: "malformed p1wins", log: "[p1wins] => nope\n[p2wins] => 2"},
		{name: "malformed p2wins", log: "[p1wins] => 1\n[p2wins] => nope"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseMatchLog(test.log); err == nil {
				t.Fatal("ParseMatchLog() error = nil, want an error")
			}
		})
	}
}
