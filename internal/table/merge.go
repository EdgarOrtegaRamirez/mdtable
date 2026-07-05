package table

import (
	"fmt"
	"strings"
)

// MergeMode defines how tables are merged.
type MergeMode int

const (
	MergeInner MergeMode = iota // Only rows with matching keys in both tables
	MergeLeft                   // All rows from left, matching from right
	MergeRight                  // All rows from right, matching from left
	MergeFull                   // All rows from both tables
)

// Merge joins two tables by a key column.
// The left table's key column index is leftKey, right table's is rightKey.
func Merge(left, right *Table, leftKey, rightKey string, mode MergeMode) (*Table, error) {
	leftIdx := left.ColumnIndex(leftKey)
	rightIdx := right.ColumnIndex(rightKey)

	if leftIdx < 0 {
		return nil, fmt.Errorf("column %q not found in left table", leftKey)
	}
	if rightIdx < 0 {
		return nil, fmt.Errorf("column %q not found in right table", rightKey)
	}

	// Build column mapping for right table (excluding the join key)
	rightCols := make(map[string]int) // header name -> original index
	for i, h := range right.Headers {
		if i == rightIdx {
			continue
		}
		// Deduplicate column names
		name := h
		if left.ColumnIndex(name) >= 0 {
			name = right.Headers[i] + "_right"
		}
		rightCols[name] = i
	}

	// Build result headers
	resultHeaders := make([]string, len(left.Headers))
	copy(resultHeaders, left.Headers)
	for name := range rightCols {
		resultHeaders = append(resultHeaders, name)
	}

	result := New(resultHeaders)

	// Copy left alignment
	for i := range left.Alignment {
		if i < len(result.Alignment) {
			result.Alignment[i] = left.Alignment[i]
		}
	}

	// Build right table lookup
	rightLookup := make(map[string][][]string) // key -> rows
	for _, row := range right.Rows {
		key := ""
		if rightIdx < len(row) {
			key = row[rightIdx]
		}
		rightLookup[key] = append(rightLookup[key], row)
	}

	// Track which right keys were matched
	matchedRight := make(map[string]bool)

	// Process left rows
	for _, leftRow := range left.Rows {
		key := ""
		if leftIdx < len(leftRow) {
			key = leftRow[leftIdx]
		}

		rightRows := rightLookup[key]
		if len(rightRows) > 0 {
			matchedRight[key] = true
			for _, rightRow := range rightRows {
				newRow := make([]string, len(resultHeaders))
				// Copy left values
				for i := range leftRow {
					newRow[i] = leftRow[i]
				}
				// Copy right values (excluding join key)
				colIdx := len(left.Headers)
				for _, origIdx := range rightCols {
					if origIdx < len(rightRow) {
						newRow[colIdx] = rightRow[origIdx]
					}
					colIdx++
				}
				result.AddRow(newRow)
			}
		} else if mode == MergeLeft || mode == MergeFull {
			newRow := make([]string, len(resultHeaders))
			for i := range leftRow {
				newRow[i] = leftRow[i]
			}
			result.AddRow(newRow)
		}
	}

	// For right/full joins, add unmatched right rows
	if mode == MergeRight || mode == MergeFull {
		for key, rightRows := range rightLookup {
			if matchedRight[key] {
				continue
			}
			for _, rightRow := range rightRows {
				newRow := make([]string, len(resultHeaders))
				// Copy right values (excluding join key)
				colIdx := len(left.Headers)
				for _, origIdx := range rightCols {
					if origIdx < len(rightRow) {
						newRow[colIdx] = rightRow[origIdx]
					}
					colIdx++
				}
				result.AddRow(newRow)
			}
		}
	}

	return result, nil
}

// Concatenate appends rows from another table to this table.
// Both tables must have the same columns (by name).
func (t *Table) Concatenate(other *Table) error {
	if len(t.Headers) != len(other.Headers) {
		return fmt.Errorf("column count mismatch: %d vs %d", len(t.Headers), len(other.Headers))
	}

	for i, h := range t.Headers {
		if strings.ToLower(h) != strings.ToLower(other.Headers[i]) {
			return fmt.Errorf("column name mismatch at position %d: %q vs %q", i, h, other.Headers[i])
		}
	}

	for _, row := range other.Rows {
		t.AddRow(row)
	}

	return nil
}
