package table

import (
	"fmt"
	"strings"
)

// DiffOp represents a single diff operation.
type DiffOp int

const (
	DiffEqual DiffOp = iota
	DiffAdded
	DiffRemoved
	DiffModified
)

// DiffRow represents a row in the diff output.
type DiffRow struct {
	Op     DiffOp
	Row    []string
	OldRow []string // only for DiffModified
}

// Diff compares two tables and returns the differences.
// It uses an LCS-based approach with modification detection.
func Diff(old, new *Table) ([]DiffRow, error) {
	if len(old.Headers) != len(new.Headers) {
		return nil, fmt.Errorf("column count mismatch: %d vs %d", len(old.Headers), len(new.Headers))
	}

	for i := range old.Headers {
		if strings.ToLower(old.Headers[i]) != strings.ToLower(new.Headers[i]) {
			return nil, fmt.Errorf("column mismatch at position %d: %q vs %q", i, old.Headers[i], new.Headers[i])
		}
	}

	// Step 1: Compute LCS to find matching rows
	lcs := computeLCS(old.Rows, new.Rows)

	// Step 2: Classify rows
	var result []DiffRow
	oldUsed := make([]bool, len(old.Rows))
	newUsed := make([]bool, len(new.Rows))

	// Mark LCS rows as equal
	for _, lcsRow := range lcs {
		for i, row := range old.Rows {
			if !oldUsed[i] && rowsEqual(row, lcsRow) {
				oldUsed[i] = true
				break
			}
		}
		for i, row := range new.Rows {
			if !newUsed[i] && rowsEqual(row, lcsRow) {
				newUsed[i] = true
				break
			}
		}
	}

	// Step 3: Try to match unmatched old rows with unmatched new rows as modifications
	type modificationPair struct {
		oldIdx int
		newIdx int
	}
	var modifications []modificationPair

	for i, used := range oldUsed {
		if used {
			continue
		}
		// Find best matching new row (by first column or position)
		bestIdx := -1
		bestScore := 0
		for j, used2 := range newUsed {
			if used2 {
				continue
			}
			score := similarityScore(old.Rows[i], new.Rows[j])
			if score > bestScore {
				bestScore = score
				bestIdx = j
			}
		}
		if bestIdx >= 0 && bestScore > 0 {
			modifications = append(modifications, modificationPair{i, bestIdx})
			oldUsed[i] = true
			newUsed[bestIdx] = true
		}
	}

	// Step 4: Build result using position-based ordering
	// Walk through both lists simultaneously, tracking positions
	oldIdx, newIdx := 0, 0

	// Build a map of modifications for quick lookup
	modMap := make(map[int]int) // oldIdx -> newIdx
	for _, m := range modifications {
		modMap[m.oldIdx] = m.newIdx
	}

	for oldIdx < len(old.Rows) || newIdx < len(new.Rows) {
		// Check if current old row has a modification
		if oldIdx < len(old.Rows) && !oldUsed[oldIdx] {
			// This old row was removed
			result = append(result, DiffRow{Op: DiffRemoved, Row: old.Rows[oldIdx]})
			oldIdx++
			continue
		}

		if newIdx < len(new.Rows) && !newUsed[newIdx] {
			// This new row was added
			result = append(result, DiffRow{Op: DiffAdded, Row: new.Rows[newIdx]})
			newIdx++
			continue
		}

		// Check for modification at oldIdx
		if newIdx2, ok := modMap[oldIdx]; ok {
			result = append(result, DiffRow{
				Op:     DiffModified,
				Row:    new.Rows[newIdx2],
				OldRow: old.Rows[oldIdx],
			})
			oldIdx++
			newIdx = newIdx2 + 1
			continue
		}

		// Both are LCS rows - they're equal
		if oldIdx < len(old.Rows) && newIdx < len(new.Rows) {
			result = append(result, DiffRow{Op: DiffEqual, Row: old.Rows[oldIdx]})
			oldIdx++
			newIdx++
		} else if oldIdx < len(old.Rows) {
			result = append(result, DiffRow{Op: DiffRemoved, Row: old.Rows[oldIdx]})
			oldIdx++
		} else {
			result = append(result, DiffRow{Op: DiffAdded, Row: new.Rows[newIdx]})
			newIdx++
		}
	}

	return result, nil
}

// similarityScore computes a similarity score between two rows (0-100).
func similarityScore(a, b []string) int {
	if len(a) == 0 && len(b) == 0 {
		return 100
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	// Count matching columns
	matches := 0
	maxCols := len(a)
	if len(b) > maxCols {
		maxCols = len(b)
	}

	for i := 0; i < maxCols; i++ {
		ai, bi := "", ""
		if i < len(a) {
			ai = strings.TrimSpace(a[i])
		}
		if i < len(b) {
			bi = strings.TrimSpace(b[i])
		}
		if ai == bi {
			matches++
		}
	}

	return (matches * 100) / maxCols
}

// computeLCS computes the Longest Common Subsequence.
func computeLCS(a, b [][]string) [][]string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if rowsEqual(a[i-1], b[j-1]) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max2(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	result := make([][]string, 0)
	i, j := m, n
	for i > 0 && j > 0 {
		if rowsEqual(a[i-1], b[j-1]) {
			result = append([][]string{a[i-1]}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return result
}

func rowsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

// RenderDiff renders a diff as a formatted string.
func RenderDiff(diffs []DiffRow, headers []string) string {
	var sb strings.Builder

	for _, d := range diffs {
		switch d.Op {
		case DiffEqual:
			sb.WriteString("  ")
			sb.WriteString(strings.Join(d.Row, " | "))
			sb.WriteString("\n")
		case DiffAdded:
			sb.WriteString("+ ")
			sb.WriteString(strings.Join(d.Row, " | "))
			sb.WriteString("\n")
		case DiffRemoved:
			sb.WriteString("- ")
			sb.WriteString(strings.Join(d.Row, " | "))
			sb.WriteString("\n")
		case DiffModified:
			sb.WriteString("- ")
			sb.WriteString(strings.Join(d.OldRow, " | "))
			sb.WriteString("\n")
			sb.WriteString("+ ")
			sb.WriteString(strings.Join(d.Row, " | "))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// DiffStats holds statistics about a diff.
type DiffStats struct {
	Added    int
	Removed  int
	Modified int
	Equal    int
}

// ComputeDiffStats computes statistics about a diff.
func ComputeDiffStats(diffs []DiffRow) DiffStats {
	var stats DiffStats
	for _, d := range diffs {
		switch d.Op {
		case DiffAdded:
			stats.Added++
		case DiffRemoved:
			stats.Removed++
		case DiffModified:
			stats.Modified++
		case DiffEqual:
			stats.Equal++
		}
	}
	return stats
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}
