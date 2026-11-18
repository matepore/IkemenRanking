# IkemenRanking

A desktop application to automate CPU-vs-CPU fights in Ikemen GO and generate
character rankings from their results.

## Project status

The application currently provides a Fyne desktop window that lets the user
select an Ikemen GO installation and its executable independently, read
characters from `data/select.def`, choose participants, configure fights per
matchup, and run an individual CPU-vs-CPU test match. Tournament automation,
ranking calculation, and CSV export are the next development steps.

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
