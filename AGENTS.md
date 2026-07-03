# AGENTS.md

## Build & Test

```bash
# Build
go build -o mdtable ./cmd/mdtable

# Run all tests (unit + integration)
go test ./...

# Run unit tests only
go test ./internal/table/ -v

# Run integration tests only
go test ./cmd/mdtable/ -v

# Lint
go vet ./...
```

## Project Structure

- `internal/table/` — Core library (parser, renderer, sort, filter, diff, merge, stats)
- `cmd/mdtable/` — CLI entry point using cobra
- Tests use table-driven patterns

## Key Design Decisions

- **No external parsing dependencies** — Markdown tables parsed with character-by-character state machine
- **No regex for table parsing** — Proper delimiter handling
- **Numeric-aware sorting** — Detects numeric columns automatically
- **LCS-based diff** — With modification detection via similarity scoring
- **Streaming stats** — No need to load all data for statistics

## Adding New Features

1. Add core logic to `internal/table/`
2. Add unit tests to `internal/table/table_test.go`
3. Add CLI command to `cmd/mdtable/main.go`
4. Add integration test to `cmd/mdtable/main_test.go`
