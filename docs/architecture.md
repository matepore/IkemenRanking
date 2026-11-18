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

The first integration step reads `data/select.def` from a user-selected
Ikemen GO installation directory. The parser:

- Reads entries only from the `[Characters]` section.
- Ignores blank lines and semicolon comments, including inline comments.
- Uses the first comma-separated field as the character identifier and ignores
  the remaining stage/order/options fields.
- Skips `randomselect` entries and duplicate identifiers (case-insensitive),
  while preserving the original spelling of the first occurrence.
- Stops when the next section begins and reports an error if `[Characters]`
  does not exist.

The parser is independent of Fyne and has unit tests. The UI displays the
identifiers exactly as they appear in the file. No Ikemen GO files are modified.

## Participant selection

The UI allows the user to choose the characters that will participate in a
future tournament:

- Every character starts unselected after an installation is loaded.
- Each character can be selected or deselected independently.
- Select-all and deselect-all controls update the full list.
- The selected-character count is displayed and remains synchronized with the
  checkboxes.
- Loading another installation clears the previous participant selection.
- Each virtualized list row uses a separate checkbox and label. The label is
  updated with the widget API so recycled rows display the correct character
  name while scrolling.

## Tournament configuration and matchup generation

The user can configure the number of individual fights played by each pair of
participants in the GUI. The default is one fight per pair. The interface
updates the estimated total fight count as the participant selection or
configured value changes.

The tournament package generates every unique unordered pair and repeats that
pair the configured number of times. It rejects fewer than two participants,
non-positive fight counts, blank participant identifiers, and duplicate
identifiers (case-insensitive). The generator is independent of Fyne and has
unit tests.

## Initial tournament rules

- Every selected character faces every other character.
- All matches are CPU versus CPU.
- A win awards one point; a loss awards zero points.
- Draws award no points.
- Match results are retained individually and exported along with the final
  standings as CSV.

These rules will be implemented in the domain packages and tested separately
from Fyne.

## Ikemen GO process integration

The first process integration step now supports one match at a time. It runs
the selected executable with the installation directory as its working
directory, requests AI level 8 for both players, uses the configured rounds
to win, and launches in windowed mode. An optional stage can be supplied.
Every invocation receives a unique temporary log filename because Ikemen GO
overwrites the specified log on subsequent runs.

The log parser reads the `p1wins`, `p2wins`, and optional `draws` counters
from Ikemen GO's Lua-table style text output. Missing or malformed required
counters are errors; they are never silently interpreted as a draw. The winner
is determined by comparing the two win counters.

## User interface

All application interface text is written in English. The UI is a presentation
layer. Long-running tournament work must run outside the UI event loop, report
progress back to the UI, and support cancellation. Domain packages must not
import Fyne.
