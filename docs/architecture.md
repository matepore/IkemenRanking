# Architecture

IkemenRanking is organized so tournament rules and Ikemen GO integration do not
depend on the graphical interface.

## Package responsibilities

- `cmd/ikemen-ranking`: application entry point.
- `internal/app`: application configuration and orchestration.
- `internal/ui`: Fyne windows, widgets, and presentation.
- `internal/ikemen`: installation validation, `select.def` parsing, process
  execution, and match-log parsing.
- `internal/tournament`: round-robin matchup generation and tournament
  execution.
- `internal/ranking`: points, wins, losses, draws, and standings.
- `internal/storage`: CSV export and any future checkpoint data.

## Character discovery

The application reads `data/select.def` from a user-selected Ikemen GO
installation directory. The parser reads entries only from the
`[Characters]` section, ignores comments and blank lines, uses the first
comma-separated field as the character identifier, skips random-select entries
and case-insensitive duplicates, and stops at the next section. It reports an
error if `[Characters]` does not exist. The stage parser reads all entries in
`[ExtraStages]`, preserving their order while ignoring comments, blank lines,
and duplicate paths. The parser is independent of Fyne and has unit tests. No
Ikemen GO files are modified.

## Participant selection

The UI allows independent character selection, select-all and deselect-all.
Loading another installation clears the previous selection. The selected
participant count is displayed. The virtualized list uses separate checkbox
and label widgets so recycled rows show the correct character.

## Tournament configuration and matchup generation

The user configures the number of individual fights played by each pair. The
default is one fight per pair, and the interface updates the estimated total
when the selection or configured value changes. The tournament package
generates each unique unordered pair the configured number of times and
validates participants and fight count.

## Single-match execution

The UI lets the user choose the Ikemen GO executable independently of its
filename, so custom builds and renamed executables are supported. The
installation directory remains the working directory, which is important for
relative game resources. A user selects two participants, AI level, rounds to
win, and chooses a stage from a dropdown populated from `[ExtraStages]`. The first
option is `Random`, which selects one of the listed stages uniformly for each
match.

The process runs outside the Fyne event loop, keeping the interface
responsive. A cancellation control terminates the launched process. The
runner gives each invocation a unique temporary log filename. The log parser
reads the `p1wins`, `p2wins`, and optional `draws` counters. Missing or
malformed required counters are errors and are not interpreted as draws.

## Initial tournament rules

- Every selected character faces every other character.
- All matches are CPU versus CPU.
- A win awards one point; a loss awards zero points.
- Draws award no points.
- Match results are retained individually and exported along with the final
  standings as CSV.

## User interface

All application interface text is written in English. The UI is a presentation
layer. Tournament orchestration and domain packages must not import Fyne.
