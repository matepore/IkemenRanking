# Architecture

IkemenRanking is organized so tournament rules and Ikemen GO integration do not
depend on the graphical interface.

## Package responsibilities

- `cmd/ikemen-ranking`: application entry point.
- `internal/app`: application configuration and orchestration.
- `internal/ui`: Fyne windows, widgets, and presentation.
- `internal/ikemen`: installation validation, `select.def` parsing, process
  execution, and match-log parsing.
- `internal/tournament`: round-robin matchup generation and sequential
  tournament execution with progress reporting.
- `internal/ranking`: points, wins, losses, draws, and standings with shared positions for tied point totals.
- `internal/storage`: persistent SQLite matchup results and global HTML/CSV ranking exports.

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

Every unique unordered participant pair is played exactly once per tournament
run. The tournament package generates each unordered pair and validates the
participant selection.

## Single-match execution

The UI lets the user choose the Ikemen GO executable independently of its
filename, so custom builds and renamed executables are supported. The
installation directory remains the working directory, which is important for
relative game resources. A user selects two participants, AI level, rounds to
win, and chooses a stage from a dropdown populated from `[ExtraStages]`. The first
option is `Random`, which selects one of the listed stages uniformly for each
match.

The process runs outside the Fyne event loop, keeping the interface
responsive. A cancellation control terminates the launched process. The same
runner is reused by the tournament executor. The
runner gives each invocation a unique temporary log filename. The log parser supports both legacy `p1wins`/`p2wins` scalar counters and the
current `wins` table (`[1]` = P1, `[2]` = P2), plus an optional `draws`
counter. Both representations are normalized to the same `MatchResult`.
Missing or malformed required counters are errors and are not interpreted as
draws.

## Tournament execution and result export

The tournament runner executes each generated fight sequentially through the
Ikemen GO match runner. It reports progress after each completed fight and
returns completed records if a later fight fails or is cancelled. The UI
exports completed records and standings calculated from them, so partial
tournament results are retained. Exports are written to a `results` directory
beside the IkemenRanking executable, created automatically when needed. Each
export updates the same self-contained, dark-themed HTML report and the same two
CSV files for fight results and standings, avoiding redundant timestamped copies. The HTML report includes the ranking, fight
scores, stages, and tournament configuration, and can be opened in a browser
without installing any additional runtime or dependencies.

The global ranking is sorted by points (one point per current matchup won, zero for a loss or draw). Each unordered matchup contributes at most one point: replaying it replaces its saved result rather than adding another point.
Characters tied on points share the same competition position; their original
participant order is retained as a stable display order.

## Initial tournament rules

- Every selected character faces every other character.
- All matches are CPU versus CPU.
- A win awards one point; a loss awards zero points.
- Draws award no points.
- The latest result for each unique unordered matchup is retained and exported alongside the global standings as CSV.

## User interface

All application interface text is written in English. The UI is a presentation
layer. Tournament orchestration and domain packages must not import Fyne.


## Persistent fight storage

The `storage.Database` component owns the local SQLite database in the application's `results` directory. It records one current result per unordered character pair, using a normalized pair key and an upsert so a replay replaces the prior result rather than accumulating attempts. Results are committed individually, allowing later tournament orchestration to recover completed fights after cancellation or application restarts. The SQLite driver is pure Go and is compiled into the application; no external SQLite DLL or installation is required.


The UI exposes a replay option (disabled by default), plus pause, resume, and cancel controls. A pause request takes effect after the active fight has completed and been persisted. Cancellation stops the active process and leaves previously committed records intact. Existing matchups are skipped unless replay is enabled; a successful replay replaces the current result for that unordered pair. The database retains one current result per pair, so a matchup can contribute at most one point to its winner. After each tournament run, the application loads every saved matchup, calculates one global ranking across all characters ever recorded, and refreshes the HTML and CSV exports from that complete dataset. Replacing a result removes the previous winner's point and awards the point to the new winner, if any. Existing database rows are retained; no tournament-history schema or data reset is required.


## Match execution diagnostics and crash retries

Failed match-log parsing preserves the generated log and includes its path in the returned error, allowing compatibility issues to be diagnosed from real engine output. The UI exposes an opt-in retry setting shared by individual test matches and tournament fights. When enabled, an unexpected non-zero Ikemen GO process exit is retried up to three additional times (four attempts total). Cancellation, invalid configuration, and match-log parsing errors are not retried.
