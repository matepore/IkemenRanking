# Architecture

IkemenRanking is organized so tournament rules and Ikemen GO integration do not
depend on the graphical interface.

## Planned package responsibilities

- `cmd/ikemen-ranking`: application entry point.
- `internal/app`: application configuration and orchestration.
- `internal/ui`: Fyne windows, widgets, and presentation.
- `internal/ikemen`: installation validation, `select.def` parsing, process
  execution, and match-log parsing.
- `internal/tournament`: round-robin matchup generation and tournament
  execution.
- `internal/ranking`: points, wins, losses, draws, and standings.
- `internal/storage`: CSV export and any future checkpoint data.

## Initial tournament rules

The initial rules agreed for the project are:

- Every selected character faces every other selected character.
- All matches are CPU versus CPU.
- A win awards one point; a loss awards zero points.
- Draws award no points.
- Match results are retained individually and exported along with the final
  standings as CSV.

These rules will be implemented in the domain packages and tested separately
from Fyne.

## Ikemen GO integration

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
