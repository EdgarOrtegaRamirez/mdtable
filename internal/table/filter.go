package table

import (
	"regexp"
	"strings"
)

// Filter keeps only rows where the given column matches the regex pattern.
func (t *Table) Filter(col int, pattern string) (*Table, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	result := t.Clone()
	result.Rows = nil

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		if re.MatchString(cell) {
			result.Rows = append(result.Rows, row)
		}
	}

	return result, nil
}

// FilterByName filters rows where a named column matches the pattern.
func (t *Table) FilterByName(name string, pattern string) (*Table, error) {
	idx := t.ColumnIndex(name)
	if idx < 0 {
		return t, nil
	}
	return t.Filter(idx, pattern)
}

// FilterNot keeps only rows where the column does NOT match the pattern.
func (t *Table) FilterNot(col int, pattern string) (*Table, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	result := t.Clone()
	result.Rows = nil

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		if !re.MatchString(cell) {
			result.Rows = append(result.Rows, row)
		}
	}

	return result, nil
}

// FilterEqual keeps rows where the column equals the value (case-insensitive).
func (t *Table) FilterEqual(col int, value string) *Table {
	result := t.Clone()
	result.Rows = nil
	lower := strings.ToLower(value)

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		if strings.ToLower(cell) == lower {
			result.Rows = append(result.Rows, row)
		}
	}

	return result
}

// FilterNotEqual keeps rows where the column does NOT equal the value.
func (t *Table) FilterNotEqual(col int, value string) *Table {
	result := t.Clone()
	result.Rows = nil
	lower := strings.ToLower(value)

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		if strings.ToLower(cell) != lower {
			result.Rows = append(result.Rows, row)
		}
	}

	return result
}

// FilterContains keeps rows where the column contains the substring.
func (t *Table) FilterContains(col int, substr string) *Table {
	result := t.Clone()
	result.Rows = nil
	lower := strings.ToLower(substr)

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		if strings.Contains(strings.ToLower(cell), lower) {
			result.Rows = append(result.Rows, row)
		}
	}

	return result
}

// FilterGreaterThan keeps rows where the column value is greater than the threshold.
func (t *Table) FilterGreaterThan(col int, threshold float64) *Table {
	result := t.Clone()
	result.Rows = nil

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		val, err := parseFloat(cell)
		if err == nil && val > threshold {
			result.Rows = append(result.Rows, row)
		}
	}

	return result
}

// FilterLessThan keeps rows where the column value is less than the threshold.
func (t *Table) FilterLessThan(col int, threshold float64) *Table {
	result := t.Clone()
	result.Rows = nil

	for _, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = row[col]
		}
		val, err := parseFloat(cell)
		if err == nil && val < threshold {
			result.Rows = append(result.Rows, row)
		}
	}

	return result
}

// Head returns the first n rows.
func (t *Table) Head(n int) *Table {
	result := t.Clone()
	if n > len(result.Rows) {
		n = len(result.Rows)
	}
	result.Rows = result.Rows[:n]
	return result
}

// Tail returns the last n rows.
func (t *Table) Tail(n int) *Table {
	result := t.Clone()
	if n > len(result.Rows) {
		n = len(result.Rows)
	}
	result.Rows = result.Rows[len(result.Rows)-n:]
	return result
}

// Unique returns only unique rows (deduplication).
func (t *Table) Unique() *Table {
	result := t.Clone()
	seen := make(map[string]bool)
	var unique [][]string

	for _, row := range result.Rows {
		key := strings.Join(row, "\x00")
		if !seen[key] {
			seen[key] = true
			unique = append(unique, row)
		}
	}

	result.Rows = unique
	return result
}

// UniqueByColumn returns only unique rows based on a specific column.
func (t *Table) UniqueByColumn(col int) *Table {
	result := t.Clone()
	seen := make(map[string]bool)
	var unique [][]string

	for _, row := range result.Rows {
		key := ""
		if col < len(row) {
			key = row[col]
		}
		if !seen[key] {
			seen[key] = true
			unique = append(unique, row)
		}
	}

	result.Rows = unique
	return result
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, &parseError{s}
	}
	// Use a simple parser instead of strconv to handle edge cases
	var result float64
	negative := false
	start := 0
	if len(s) > 0 && s[0] == '-' {
		negative = true
		start = 1
	} else if len(s) > 0 && s[0] == '+' {
		start = 1
	}

	foundDot := false
	multiplier := 1.0

	for i := start; i < len(s); i++ {
		ch := s[i]
		if ch == '.' {
			if foundDot {
				return 0, &parseError{s}
			}
			foundDot = true
			continue
		}
		if ch < '0' || ch > '9' {
			return 0, &parseError{s}
		}
		digit := float64(ch - '0')
		if foundDot {
			multiplier *= 0.1
			result += digit * multiplier
		} else {
			result = result*10 + digit
		}
	}

	if negative {
		result = -result
	}
	return result, nil
}

type parseError struct {
	input string
}

func (e *parseError) Error() string {
	return "not a number: " + e.input
}
