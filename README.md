# IkemenRanking

A desktop application to automate CPU-vs-CPU fights in Ikemen GO and generate
character rankings from their results.

## Project status

The repository currently contains the initial Go/Fyne desktop application
skeleton. Tournament execution, character-list parsing, result processing, and
CSV export are planned next.

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

See [docs/architecture.md](docs/architecture.md) for the planned package
boundaries and integration approach.
