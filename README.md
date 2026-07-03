# mdtable

A fast, comprehensive markdown table toolkit for the command line. Parse, sort, filter, merge, diff, and convert markdown tables between formats.

## Why?

Markdown tables are everywhere — READMEs, documentation, reports, data dumps. But there's no good CLI tool for *manipulating* them. `csv2md` only converts CSV→Markdown. `jtbl` only converts JSON→tables. `mdtable` works **bidirectionally** with tables you already have.

## Features

- **Parse** markdown tables from files or stdin
- **Sort** by any column (numeric-aware, ascending/descending)
- **Filter** rows by regex, equality, contains, greater/less than
- **Head/Tail** — show first/last N rows
- **Unique** — deduplicate rows (whole row or by column)
- **Merge** — join two tables by key column (inner, left, right, full)
- **Diff** — compare two tables and show additions, removals, modifications
- **Convert** — output as CSV, JSON, HTML, TSV, or markdown
- **Stats** — compute statistics (min, max, mean, median, stddev, unique count)
- **Columns** — list column names, types, and alignment
- **Add/Remove columns** — programmatically manipulate table structure

## Install

```bash
go install github.com/EdgarOrtegaRamirez/mdtable/cmd/mdtable@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/mdtable.git
cd mdtable
go build -o mdtable ./cmd/mdtable
```

## Quick Start

```bash
# Sort a table by score
cat scores.md | mdtable sort Score

# Sort descending
cat scores.md | mdtable sort Score -d

# Filter rows matching a pattern
cat scores.md | mdtable filter City "^N"

# Show first 5 rows
cat data.md | mdtable head 5

# Convert to CSV
cat data.md | mdtable convert csv

# Convert to JSON
cat data.md | mdtable convert json

# Show column statistics
cat data.md | mdtable stats

# Diff two versions of a table
mdtable diff old-table.md < new-table.md
```

## Commands

| Command | Description |
|---------|-------------|
| `sort [col]` | Sort by column (numeric-aware) |
| `filter [col] [pattern]` | Filter rows by regex |
| `head [n]` | Show first N rows |
| `tail [n]` | Show last N rows |
| `unique [col]` | Remove duplicate rows |
| `merge [file]` | Join two tables by key |
| `diff [file]` | Compare two tables |
| `convert [fmt]` | Convert to csv/json/html/tsv |
| `stats` | Column statistics |
| `columns` | List column metadata |

## Examples

### Sort and filter

```bash
# Sort employees by department, then salary descending
cat employees.md | mdtable sort Department | mdtable sort Score -d
```

### Merge two tables

```bash
# Join users and orders by user ID
mdtable merge orders.md -k ID < users.md
```

### Generate CSV for spreadsheets

```bash
cat report.md | mdtable convert csv > report.csv
```

### Compare table versions

```bash
mdtable diff before.md < after.md
```

Output:
```
- Alice | 90 | NYC
+ Alice | 95 | NYC
  Bob   | 75 | London
+ Dave  | 80 | Berlin
```

## Architecture

```
mdtable/
├── internal/table/     # Core library
│   ├── table.go       # Data structures, column operations
│   ├── parse.go       # Markdown/CSV/JSON parsers
│   ├── render.go      # Markdown/CSV/JSON/HTML/TSV renderers
│   ├── sort.go        # Numeric-aware sorting (multi-key)
│   ├── filter.go      # Regex, equality, comparison filters
│   ├── diff.go        # LCS-based diff with modification detection
│   ├── merge.go       # Inner/left/right/full joins
│   └── stats.go       # Statistical analysis
├── cmd/mdtable/       # CLI (cobra)
└── tests/             # Integration tests
```

### Key Algorithms

- **Parsing**: Character-by-character state machine (not regex)
- **Sorting**: Go's `sort.SliceStable` with numeric-aware comparison
- **Diff**: LCS (Longest Common Subsequence) with modification detection via similarity scoring
- **Merge**: Hash-join for O(n+m) performance
- **Statistics**: Streaming computation (no full materialization needed)

## Input Formats

mdtable reads:
- Markdown tables (with `|` delimiters)
- Tables without leading/trailing pipes
- Standard markdown with `---` separator rows

## License

MIT
