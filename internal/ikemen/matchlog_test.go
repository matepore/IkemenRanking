package ikemen

import "testing"

func TestParseMatchLogLegacyFormat(t *testing.T) {
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

func TestParseMatchLogCurrentWinsTableFormat(t *testing.T) {
	contents := `{
		["winSide"] => 1,
		["lastRound"] => 3,
		["draws"] => 1,
		["wins"] => table: 0x123456 {
			[1] => 1
			[2] => 2
		},
		["rounds"] => table: 0x654321 {
			[1] => table: 0x111111 {
				["score"] => table: 0x222222 {
					[1] => 1
					[2] => 0
				}
			}
		}
	}`

	got, err := ParseMatchLog(contents)
	if err != nil {
		t.Fatalf("ParseMatchLog() error = %v", err)
	}
	if got.Player1Wins != 1 || got.Player2Wins != 2 || got.Draws != 1 {
		t.Fatalf("unexpected counters: %#v", got)
	}
}

func TestParseMatchLogCurrentWinsTableAllowsMissingDrawCounter(t *testing.T) {
	got, err := ParseMatchLog(`["wins"] => table: 0x123 {
		[1] => 2
		[2] => 1
	}`)
	if err != nil {
		t.Fatalf("ParseMatchLog() error = %v", err)
	}
	if got.Player1Wins != 2 || got.Player2Wins != 1 || got.Draws != 0 {
		t.Fatalf("unexpected counters: %#v", got)
	}
}

func TestParseMatchLogLegacyFormatAllowsMissingDrawCounter(t *testing.T) {
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
		{name: "missing legacy p1wins", log: "[p2wins] => 2"},
		{name: "missing legacy p2wins", log: "[p1wins] => 2"},
		{name: "malformed legacy p1wins", log: "[p1wins] => nope\n[p2wins] => 2"},
		{name: "malformed legacy p2wins", log: "[p1wins] => 1\n[p2wins] => nope"},
		{name: "missing current P1 counter", log: `["wins"] => table: 0x123 {
			[2] => 2
		}`},
		{name: "missing current P2 counter", log: `["wins"] => table: 0x123 {
			[1] => 2
		}`},
		{name: "malformed current counter", log: `["wins"] => table: 0x123 {
			[1] => nope
			[2] => 2
		}`},
		{name: "unterminated current table", log: `["wins"] => table: 0x123 {
			[1] => 1
			[2] => 2`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseMatchLog(test.log); err == nil {
				t.Fatal("ParseMatchLog() error = nil, want an error")
			}
		})
	}
}

func TestParseMatchLogDoesNotReadRoundScoresAsMatchWins(t *testing.T) {
	contents := `["rounds"] => table: 0x1 {
		[1] => table: 0x2 {
			["score"] => table: 0x3 {
				[1] => 2
				[2] => 0
			}
		}
	}`

	if _, err := ParseMatchLog(contents); err == nil {
		t.Fatal("ParseMatchLog() error = nil, want an error")
	}
}
