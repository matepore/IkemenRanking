# IkemenRanking

A desktop application to automate CPU-vs-CPU fights in Ikemen GO and generate
character rankings from their results.

## Project status

The application stores completed fight results in a local SQLite database under `results/`. The database is embedded through a pure-Go SQLite driver, so end users do not need to install SQLite separately.

The application provides a Fyne desktop window to select an Ikemen GO
installation and executable, read characters and stages from `data/select.def`,
choose participants, run an individual test match, and execute a round-robin
CPU-vs-CPU tournament with one fight per unique pair. Tournament progress and
standings are displayed in the UI. Each tournament exports a standalone dark
HTML report plus separate fight-results and standings CSV files to a
`results` directory beside the IkemenRanking executable.

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

Build an executable:

```sh
go build -o bin/ikemen-ranking ./cmd/ikemen-ranking
```

## Documentation

See [docs/architecture.md](docs/architecture.md) for the package boundaries
and integration approach.
