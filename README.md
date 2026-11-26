# IkemenRanking

A desktop application to automate CPU-vs-CPU fights in Ikemen GO and generate
character rankings from their results.

## Project status

The application stores completed fight results in a local SQLite database under `results/`. The database is embedded through a pure-Go SQLite driver, so end users do not need to install SQLite separately.

The application provides a Fyne desktop window to select an Ikemen GO
installation and executable, read characters and stages from `data/select.def`,
choose participants, and execute a round-robin CPU-vs-CPU tournament with one
fight per unique pair. Tournament progress and standings are displayed in the
UI. Each tournament exports a standalone dark HTML report plus separate
fight-results and standings CSV files to a `results` directory beside the
IkemenRanking executable. Installation paths can be entered directly or
selected with **Browse**; invalid paths are reported inline.

## Development prerequisites

- Go 1.22 or newer.
- A C compiler and the platform development dependencies required by Fyne.
- On Windows, a working MinGW-w64 toolchain is the recommended starting point
  for native Fyne builds.

These are development/build requirements. The goal for the end-user release is
a self-contained executable that does not require Go or a separately installed
runtime.

## Run locally

From the repository root:

```sh
go mod tidy
go run ./cmd/ikemen-ranking
```

Build an executable (without the Windows file-resource icon):

```sh
go build -o bin/ikemen-ranking.exe ./cmd/ikemen-ranking
```

For the Windows executable file icon, install the current Fyne CLI and run the
packaging command from the application entry-point directory:

```powershell
go install fyne.io/tools/cmd/fyne@latest
Push-Location .\cmd\ikemen-ranking
fyne package -os windows -icon "..\..\icon\icon(256x256).png"
Pop-Location
```

The 256x256 PNG is embedded and assigned to both the Fyne application and
window at runtime. The same PNG is embedded into the Windows executable
resource by the packaging command and is used for the icon shown by Explorer.
A plain `go build` does not add that Windows executable resource.

## Documentation

- [Architecture](docs/architecture.md): package boundaries and integration
  approach.
- [Project progress and handoff](PROJECT_PROGRESS.md): detailed project state,
  decisions, milestones, verification history, and continuation instructions.

Tournament execution can be paused safely between fights and resumed in the same session. Cancelling a tournament preserves every result already committed to SQLite. On a later launch, running the same participant selection resumes by skipping previously completed matchups. Enable **Replay previously completed fights** to run those pairs again; each successful replay replaces its prior result.

## License

IkemenRanking is distributed under the MIT License. See the [LICENSE](LICENSE)
file for the complete license text.
