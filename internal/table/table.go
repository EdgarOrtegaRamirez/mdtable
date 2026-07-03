// Package table provides data structures and operations for markdown tables.
//
// A markdown table consists of:
//   - Header row: the first row with column names
//   - Separator row: alignment indicators (---, :---:, ---:)
//   - Data rows: the actual data
//
// Tables support bidirectional conversion with CSV, JSON, HTML, and TSV formats.
package table

import (
	"fmt"
	"strings"
)

// Alignment represents column text alignment.
type Alignment int

const (
	AlignLeft   Alignment = iota // Default
	AlignCenter                  // :---:
	AlignRight                   // ---:
)

// String returns the alignment marker for markdown output.
func (a Alignment) String() string {
	switch a {
	case AlignCenter:
		return ":---:"
	case AlignRight:
		return "---:"
	default:
		return "---"
	}
}

// ParseAlignment parses an alignment marker string.
func ParseAlignment(s string) Alignment {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return AlignLeft
	}
	leftColon := s[0] == ':'
	rightColon := s[len(s)-1] == ':'
	if leftColon && rightColon {
		return AlignCenter
	}
	if rightColon {
		return AlignRight
	}
	return AlignLeft
}

// Table represents a structured markdown table.
type Table struct {
	Headers   []string
	Alignment []Alignment
	Rows      [][]string
}

// New creates a new empty table with the given headers.
func New(headers []string) *Table {
	align := make([]Alignment, len(headers))
	for i := range align {
		align[i] = AlignLeft
	}
	return &Table{
		Headers:   headers,
		Alignment: align,
		Rows:      make([][]string, 0),
	}
}

// NumCols returns the number of columns.
func (t *Table) NumCols() int {
	return len(t.Headers)
}

// NumRows returns the number of data rows.
func (t *Table) NumRows() int {
	return len(t.Rows)
}

// AddRow adds a row to the table. If the row has fewer columns than headers,
// it is padded with empty strings. If it has more, excess columns are truncated.
func (t *Table) AddRow(row []string) {
	padded := make([]string, len(t.Headers))
	for i := range padded {
		if i < len(row) {
			padded[i] = strings.TrimSpace(row[i])
		}
	}
	t.Rows = append(t.Rows, padded)
}

// ColumnIndex returns the index of a column by name, or -1 if not found.
// Comparison is case-insensitive.
func (t *Table) ColumnIndex(name string) int {
	lower := strings.ToLower(name)
	for i, h := range t.Headers {
		if strings.ToLower(h) == lower {
			return i
		}
	}
	return -1
}

// ColumnValues returns all values in a column.
func (t *Table) ColumnValues(col int) []string {
	vals := make([]string, len(t.Rows))
	for i, row := range t.Rows {
		if col < len(row) {
			vals[i] = row[col]
		}
	}
	return vals
}

// ColumnValuesByName returns all values in a named column.
func (t *Table) ColumnValuesByName(name string) ([]string, bool) {
	idx := t.ColumnIndex(name)
	if idx < 0 {
		return nil, false
	}
	return t.ColumnValues(idx), true
}

// IsNumericColumn checks if all non-empty values in a column are numeric.
func (t *Table) IsNumericColumn(col int) bool {
	count := 0
	for _, row := range t.Rows {
		if col >= len(row) {
			continue
		}
		v := strings.TrimSpace(row[col])
		if v == "" {
			continue
		}
		count++
		if !isNumeric(v) {
			return false
		}
	}
	return count > 0
}

// Clone returns a deep copy of the table.
func (t *Table) Clone() *Table {
	c := &Table{
		Headers:   make([]string, len(t.Headers)),
		Alignment: make([]Alignment, len(t.Alignment)),
		Rows:      make([][]string, len(t.Rows)),
	}
	copy(c.Headers, t.Headers)
	copy(c.Alignment, t.Alignment)
	for i, row := range t.Rows {
		c.Rows[i] = make([]string, len(row))
		copy(c.Rows[i], row)
	}
	return c
}

// Subset returns a new table with only the specified columns.
func (t *Table) Subset(cols []int) *Table {
	c := &Table{
		Headers:   make([]string, len(cols)),
		Alignment: make([]Alignment, len(cols)),
		Rows:      make([][]string, len(t.Rows)),
	}
	for i, col := range cols {
		if col < len(t.Headers) {
			c.Headers[i] = t.Headers[col]
		}
		if col < len(t.Alignment) {
			c.Alignment[i] = t.Alignment[col]
		}
	}
	for i, row := range t.Rows {
		c.Rows[i] = make([]string, len(cols))
		for j, col := range cols {
			if col < len(row) {
				c.Rows[i][j] = row[col]
			}
		}
	}
	return c
}

// SwapRows swaps two rows by index.
func (t *Table) SwapRows(i, j int) {
	t.Rows[i], t.Rows[j] = t.Rows[j], t.Rows[i]
}

// AddColumn adds a new column with the given name and default values.
func (t *Table) AddColumn(name string, values []string) {
	t.Headers = append(t.Headers, name)
	t.Alignment = append(t.Alignment, AlignLeft)
	for i := range t.Rows {
		if i < len(values) {
			t.Rows[i] = append(t.Rows[i], values[i])
		} else {
			t.Rows[i] = append(t.Rows[i], "")
		}
	}
	// Pad existing rows if values has more entries
	for len(values) > len(t.Rows) {
		row := make([]string, len(t.Headers))
		idx := len(t.Rows)
		row[len(t.Headers)-1] = values[idx]
		t.Rows = append(t.Rows, row)
	}
}

// RemoveColumn removes a column by index.
func (t *Table) RemoveColumn(col int) {
	if col < 0 || col >= len(t.Headers) {
		return
	}
	t.Headers = append(t.Headers[:col], t.Headers[col+1:]...)
	t.Alignment = append(t.Alignment[:col], t.Alignment[col+1:]...)
	for i := range t.Rows {
		if col < len(t.Rows[i]) {
			t.Rows[i] = append(t.Rows[i][:col], t.Rows[i][col+1:]...)
		}
	}
}

// RenameColumn renames a column.
func (t *Table) RenameColumn(col int, name string) {
	if col >= 0 && col < len(t.Headers) {
		t.Headers[col] = name
	}
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// Handle negative numbers
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	hasDot := false
	for _, c := range s {
		if c == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// FormatError is returned when table parsing fails.
type FormatError struct {
	Line    int
	Message string
}

func (e *FormatError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}
