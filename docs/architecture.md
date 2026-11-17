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

The selection is presentation state for now; tournament generation will
consume the selected identifiers in a later implementation step.

## Initial tournament rules

- Every selected character faces every other selected character.
- All matches are CPU versus CPU.
- A win awards one point; a loss awards zero points.
- Draws award no points.
- Match results are retained individually and exported along with the final
  standings as CSV.

These rules will be implemented in the domain packages and tested separately
from Fyne.

## Ikemen GO process integration

The application will launch the user's existing Ikemen GO executable as an
external process. The game installation and its files are not bundled with
IkemenRanking, and the project will avoid requiring modifications to the game
source code.

Each match should use its own log file so a later process run cannot overwrite
an earlier result. Missing, malformed, or incomplete logs must be treated as
an unknown/failed match rather than being interpreted as a win or draw.

## User interface

The UI is a presentation layer. Long-running tournament work must run outside
the UI event loop, report progress back to the UI, and support cancellation.
Domain packages must not import Fyne.
