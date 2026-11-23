# IkemenRanking — Project Progress & Handoff

> Living project record. Update this document in the same commit as every
> meaningful code, configuration, test, or documentation change. Its purpose is
> to let work continue in a new ChatGPT conversation without relying on the
> previous chat transcript.

## 1. Project overview

**IkemenRanking** is a Windows desktop application that automates CPU-versus-CPU
fights in Ikemen GO and builds a persistent global character ranking from the
results.

The intended end-user experience is a portable Windows executable. End users
should not need to install Go, Python, SQLite, or another runtime. The
application uses Fyne for its desktop UI and a pure-Go SQLite driver for
persistence.

- Repository: https://github.com/matepore/IkemenRanking
- Default integration branch: `main`
- Main language: Go (module targets Go 1.22)
- UI language: English
- Primary development environment: Windows / PowerShell
- Local working directory: `C:\Otros\Repositorios\IkemenRanking`

## 2. Product goals and constraints

1. Discover a user's Ikemen GO installation and executable.
2. Read available characters and stages from the installation's
   `data/select.def`.
3. Let the user select two characters for an individual test fight.
4. Run CPU-versus-CPU matches with configurable AI level, rounds to win,
   stage, and game speed.
5. Run sequential round-robin tournaments over selected characters.
6. Persist results locally so rankings survive application restarts and
   subsequent tournament runs.
7. Maintain one cumulative global ranking, not separate tournament rankings.
8. Export a self-contained HTML report and CSV data.
9. Support multiple Ikemen GO generations where their command-line or log
   interfaces differ.
10. Keep the UI in English and keep game integration/domain logic independent
    from Fyne.

## 3. Technology and package architecture

| Area | Implementation |
|---|---|
| Language | Go 1.22+ |
| Desktop UI | Fyne v2.6.3 |
| Database | SQLite via `modernc.org/sqlite` v1.34.5 (pure Go) |
| Application entry point | `cmd/ikemen-ranking` |
| App orchestration | `internal/app` |
| UI | `internal/ui` |
| Ikemen GO integration | `internal/ikemen` |
| Tournament generation/execution | `internal/tournament` |
| Ranking calculations | `internal/ranking` |
| Persistence and exports | `internal/storage` |

The UI is a presentation layer. Tournament rules, ranking calculations,
persistence, and Ikemen GO integration should remain testable without importing
Fyne.

See [docs/architecture.md](docs/architecture.md) for the current architecture
notes.

## 4. User-facing functionality implemented

### Installation and content discovery

- Select an Ikemen GO installation directory and executable independently of
  the executable's filename.
- Keep the installation directory as the launched process working directory so
  the engine can resolve its relative resources.
- Read the `[Characters]` section of `data/select.def`.
- Ignore comments, blank lines, random-select entries, and case-insensitive
  duplicate character identifiers.
- Read stage entries from `[ExtraStages]`, preserving order and removing
  duplicates.
- Do not modify the user's Ikemen GO installation files.

### Match and tournament execution

- Individual CPU-vs-CPU match configuration.
- Configurable AI level and rounds to win.
- Stage selection, including a Random option.
- Configurable game speed, with the available options adapted to the detected
  Ikemen GO command-line convention.
- Sequential round-robin tournament: each unique unordered pair is played once
  per tournament run.
- Tournament progress reporting.
- Pause, resume, and cancel controls. Pause takes effect between fights;
  cancelling preserves already persisted results.
- Previously completed matchups are skipped by default. A replay option allows
  them to be run again; a successful replay replaces the stored result.

### Global ranking rules

The ranking is global across all characters that have participated in any
saved fight. It is not reset between tournaments.

- One point for the current winning result of a unique unordered matchup.
- A loss awards no points.
- A draw awards no points.
- Each unordered pair contributes at most one point in total.
- Replaying a matchup replaces its previous result. If the winner changes, the
  previous winner loses that point and the new winner receives it.
- Replaying a matchup with the same winner does not add another point.
- A draw means neither character receives a point for that matchup.
- Characters tied on points share the same competition position. Original
  participant order is retained as a stable display order.

### Persistence and exports

SQLite is the source of truth. The database is stored under `results/` beside
the application executable, using the filename `ikemen-ranking.db`.

Exports are also written under `results/`:

- `ikemen-ranking-report.html`
- `ikemen-ranking-matches.csv`
- `ikemen-ranking-standings.csv`

The HTML report is self-contained and uses a dark esports-style design. The
global ranking and exports are refreshed after completed fights so saved
results are reflected without waiting for another tournament run.

## 5. Ikemen GO compatibility

### Game speed conventions

Ikemen GO has used two different `-speed` conventions:

| Engine generation | Normal speed | Faster options | Detection |
|---|---:|---|---|
| Legacy releases (e.g. v0.99.0) | 100% | 110% through 200% | Inspect `external/script/main.lua` for the percentage-based implementation |
| v1.0.0+ / current develop | 0 | +1 through +9 | Inspect `external/script/main.lua` for the step-based implementation |

The application inspects the actual Lua implementation instead of trusting only
version metadata. This supports custom builds whose version string may not
reliably identify the implementation.

### Match-log formats

**Legacy format:** separate scalar fields:

```text
[p1wins] => 0
[p2wins] => 2
[draws] => 0
```

**Current format:** a `Wins` table, where array index 1 is P1 and index 2 is P2. Real Nightly output capitalizes keys such as `Draws` and `Wins`; the parser matches field names case-insensitively:

```text
["draws"] => 1
["wins"] => table: 0x123456 {
  [1] => 1
  [2] => 2
}
```

The current format follows the upstream `StatsMatch.Wins [2]int32` definition
and the Lua `main.f_printTable(getGameStats().Matches[matchNo()], ...)` log
writer. The parser normalizes both formats to the same internal
`MatchResult`. It reads indices only from the `wins` table, not from nested
round score tables. Missing or malformed required win counters remain errors;
they must never be silently treated as draws.

Relevant upstream implementation references:

- https://github.com/ikemen-engine/Ikemen-GO/blob/develop/src/stats.go
- https://github.com/ikemen-engine/Ikemen-GO/blob/develop/src/script.go
- https://github.com/ikemen-engine/Ikemen-GO/blob/develop/external/script/main.lua

## 6. Important implementation decisions

- Detect engine behavior from the installed scripts where practical, rather
  than relying solely on version labels.
- Keep version-specific parsing/adaptation inside `internal/ikemen`.
- Normalize engine-specific results into a stable application-level model.
- Do not make tournament, ranking, storage, or UI code understand individual
  Ikemen GO log formats.
- Preserve strict validation: unknown or incomplete formats should produce a
  useful error instead of corrupting ranking data.
- SQLite stores the latest result for each normalized unordered pair; it does
  not store tournament history.
- Keep the portable end-user distribution free from external runtime
  dependencies.

## 7. Manual verification reported by the user

The user has manually confirmed the following behavior:

- Installation discovery and loading characters/stages work.
- Character selection works.
- Individual CPU-vs-CPU matches work with the user's legacy Ikemen GO version.
- Sequential tournament execution works.
- SQLite persistence survives application restarts.
- The global ranking is correct and cumulative across tournament runs.
- HTML and CSV exports work.
- The HTML report's visual design is satisfactory.
- The UI scrolling and fixed-height character selection behavior work.
- Legacy speed detection and speed selection work.
- Current Nightly speed detection and speed selection work.

**Nightly match-log issue:** the fight completed but the previous parser failed
because it required legacy `p1wins` / `p2wins` fields. The parser has now been
extended to accept the current `wins[1]` / `wins[2]` representation as well.
Automated tests were added for both formats and malformed/incomplete data.
The user supplied the real Nightly log. It uses capitalized keys `Draws` and `Wins`, with `Wins[1] = 2` and `Wins[2] = 0`. The earlier parser expected lowercase key names, explaining why it did not recognize the table. The parser now matches field names case-insensitively, with a regression test for Nightly capitalization and nested round data. A prior change also preserves failed logs and adds an opt-in checkbox to retry unexpected engine process exits up to three times; parser errors and cancellations are not retried.

## 8. Test and build workflow

Run the complete Go test suite from the repository root:

```powershell
go test ./...
```

Previously, the user reported that `go test ./...` passed across all packages
after the global-ranking and speed-compatibility work. The current match-log
compatibility changes have not yet been executed in a local Go environment by
the assistant; do not assume the new tests pass until run.

Build command:

```powershell
go build -o bin/ikemen-ranking.exe ./cmd/ikemen-ranking
```

Do not claim a test or build passed unless it was actually run. Automated tests
cannot replace the user's manual verification against their installed legacy
and Nightly engine builds.

## 9. Commit history and milestones

The following commits are known from the project work:

| Commit | Change |
|---|---|
| `5d5fa269938dccfe12ae51d7ce347ed68f972e15` | Global ranking based on all saved matchups |
| `eba2a0b23d8c280f3da8ac4f453f37ce238a590e` | Refresh global ranking after every completed fight |
| `095e8c2155755bd10cc1a8576efa4faea8b2b8c0` | Detect Ikemen GO game speed compatibility |
| `bb64849a2d6a295c117ea174a8aa7a908b6cbb76` | Support legacy and current Ikemen GO match-log formats; add regression tests and project handoff documentation |
| `a5bcc727283664d90976cf6d28719317492ffa83` | Correct Go/RE2 regular-expression syntax in the match-log parser |
| `bbcf128eade5ce0e6a325050cf4843e4f31c77ce` | Fix capture-group selection for the optional draws counter |

**Follow-up after user's test report:** the first user-run `go test ./...`
identified a capture-group indexing bug in the optional `draws` field: after
making the expression Go/RE2-compatible, the key name was capture group 1 and
the numeric value was capture group 2. The parser now converts the final capture
group for that expression, and a regression test covers both quoted and
unquoted `draws` keys. The regression fixtures use actual newline characters
so the multiline parser behavior is exercised. This correction is committed
but still requires a fresh local `go test ./...` run from the user.

## 10. Current status and next steps

### Immediate next steps

1. Build and run `go test ./...` after the diagnostic-log and crash-retry changes.
2. Run one Nightly fight. If parsing fails, the error will include the preserved log path; share that file so the parser can be fixed against the actual output.
3. Verify that enabling the crash-retry checkbox retries an unexpected non-zero Ikemen GO process exit, up to three additional attempts.
4. Confirm that disabling the checkbox preserves the previous fail-fast behavior, and that parser errors are not retried.
5. Run a legacy-engine fight to check for regressions.

### Compatibility audit to follow

The user wants the application to accommodate behavioral changes across
Ikemen GO development versions. Continue auditing integration points
systematically, prioritizing evidence-backed differences:

1. Match result/log output (current change).
2. Command-line argument names and semantics.
3. Character and stage selection conventions.
4. Match configuration and game launch behavior.
5. Any future engine-side API changes that affect the app.

Do not speculate that a particular interface changed without checking the
relevant source or reproducing the behavior. Prefer capability/format detection
over version checks when the implementation can be inspected reliably.

## 11. Handoff instructions for a future conversation

When continuing this project in a new chat:

1. Read this entire `PROJECT_PROGRESS.md` first.
2. Work only on `feature/initial-gui`; do not commit to `main`.
3. Fetch the current branch head before any GitHub write. Never overwrite
   changes made since the last known commit.
4. Inspect the current repository files before editing; this document is a
   handoff aid, not a substitute for source code.
5. Preserve the global ranking rules and existing SQLite data model unless the
   user explicitly changes the requirements.
6. Keep the UI text in English.
7. Update this document in the same commit as every code/configuration/test/doc
   change. Record the commit SHA, summary, files changed, and actual test/build
   results. Clearly distinguish automated verification from manual checks
   performed by the user. Because a commit cannot include its own SHA, record
   its SHA in the next progress-document update.
8. After a commit, verify the branch ref and inspect the resulting commit diff.

This file should remain concise enough to navigate but detailed enough to
recover decisions, constraints, current state, and the next concrete actions.


### Pending change in this work session

- Preserve the generated match log when reading or parsing fails and include its path in the error, instead of deleting the evidence unconditionally.
- Add a checkbox shared by single-match tests and tournaments to retry unexpected non-zero Ikemen GO process exits up to three times (four total attempts). Intermediate crash logs are discarded; the final failed attempt keeps its log. Match-log parsing errors and cancellation are not retried.
- This change is not considered verified until the user runs the full Go test suite and performs a Nightly test.


### Nightly log fixture follow-up

- Actual user-provided Nightly output uses capitalized `Draws` and `Wins` keys. The previous parser's case-sensitive lowercase patterns missed the `Wins` table and incorrectly fell back to the legacy `p1wins` error.
- Updated key matching to be case-insensitive and added a regression test with the real capitalization, `Wins[1] = 2`, `Wins[2] = 0`, and nested round score data.
- Expected parsed result for the supplied log: P1 wins 2, P2 wins 0, draws 0.


## 12. Graphite & Crimson UI refresh

- Approved visual direction: Graphite & Crimson, balanced crimson accent.
- Implementation branch: `feature/graphite-theme`, created from `main` at `7e1d703062fcaee104988e3790a93c886166aa03`.
- Added a custom Fyne theme with graphite backgrounds, muted text, crimson primary/focus states, and semantic success/warning/error colors.
- Added a reusable section-card helper and grouped the main window into distinct installation, character-selection, character-roster, match-configuration, tournament-settings, and tournament-execution panels.
- Applied the custom theme at application startup. The existing vertically scrollable single-window workflow is retained.
- No tournament, ranking, persistence, export, or Ikemen GO integration logic was intentionally changed.
- Automated tests and Windows GUI build have **not** been run by the assistant in this environment. User-side visual verification is still required, particularly for card spacing, scroll behavior, and long character lists.
- Commits on this branch:
  - `b3cbb775cc3639a00480f33bef53d57fd3fcfb2f` — add Graphite & Crimson theme.
  - `711b97dbe2009c62094bf04d05ae5f1918780ddc` — add reusable section cards.
  - `10d3769fd51c98ac384dbfb33ed71ae664235d98` — apply theme at startup.
  - `b7c89004d57d3de264f5fc6810c7d566e8b4504c` — reorganize main window into section panels.
