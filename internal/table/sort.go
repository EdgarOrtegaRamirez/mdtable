package table

import (
	"sort"
	"strconv"
	"strings"
)

// SortByColumn sorts the table by a column index.
// For numeric columns, it sorts numerically; otherwise lexicographically.
func (t *Table) SortByColumn(col int, descending bool) {
	if col < 0 || col >= len(t.Headers) {
		return
	}

	numeric := t.IsNumericColumn(col)

	sort.SliceStable(t.Rows, func(i, j int) bool {
		a := ""
		if col < len(t.Rows[i]) {
			a = t.Rows[i][col]
		}
		b := ""
		if col < len(t.Rows[j]) {
			b = t.Rows[j][col]
		}

		var cmp int
		if numeric {
			cmp = compareNumeric(a, b)
		} else {
			cmp = strings.Compare(strings.ToLower(a), strings.ToLower(b))
		}

		if descending {
			return cmp > 0
		}
		return cmp < 0
	})
}

// SortByColumnName sorts the table by a named column.
func (t *Table) SortByColumnName(name string, descending bool) {
	idx := t.ColumnIndex(name)
	if idx >= 0 {
		t.SortByColumn(idx, descending)
	}
}

// compareNumeric compares two numeric strings.
func compareNumeric(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	if a == "" && b == "" {
		return 0
	}
	if a == "" {
		return -1
	}
	if b == "" {
		return 1
	}

	af, errA := strconv.ParseFloat(a, 64)
	bf, errB := strconv.ParseFloat(b, 64)

	if errA != nil && errB != nil {
		return strings.Compare(a, b)
	}
	if errA != nil {
		return -1
	}
	if errB != nil {
		return 1
	}

	if af < bf {
		return -1
	}
	if af > bf {
		return 1
	}
	return 0
}

// MultiSort sorts by multiple columns in priority order.
// Each entry specifies a column index and direction.
type SortKey struct {
	Column     int
	Descending bool
}

// SortMulti sorts the table by multiple columns.
func (t *Table) SortMulti(keys []SortKey) {
	// Precompute numeric flags
	numericFlags := make([]bool, len(keys))
	for i, k := range keys {
		numericFlags[i] = t.IsNumericColumn(k.Column)
	}

	sort.SliceStable(t.Rows, func(i, j int) bool {
		for idx, key := range keys {
			col := key.Column
			a := ""
			if col < len(t.Rows[i]) {
				a = t.Rows[i][col]
			}
			b := ""
			if col < len(t.Rows[j]) {
				b = t.Rows[j][col]
			}

			var cmp int
			if numericFlags[idx] {
				cmp = compareNumeric(a, b)
			} else {
				cmp = strings.Compare(strings.ToLower(a), strings.ToLower(b))
			}

			if cmp == 0 {
				continue
			}

			if key.Descending {
				return cmp > 0
			}
			return cmp < 0
		}
		return false // all equal
	})
}
