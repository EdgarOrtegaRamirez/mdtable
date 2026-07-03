package table

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Stats holds statistical summary for a column.
type Stats struct {
	Column  string
	Count   int
	Min     string
	Max     string
	Sum     float64
	Mean    float64
	Median  float64
	StdDev  float64
	Unique  int
	Empty   int
	IsNum   bool
}

// ColumnStats computes statistics for a column.
func (t *Table) ColumnStats(col int) Stats {
	if col < 0 || col >= len(t.Headers) {
		return Stats{}
	}

	stats := Stats{
		Column: t.Headers[col],
		IsNum:  t.IsNumericColumn(col),
	}

	values := t.ColumnValues(col)
	uniqueSet := make(map[string]bool)

	var numValues []float64
	var minVal, maxVal string
	first := true

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			stats.Empty++
			continue
		}

		stats.Count++
		uniqueSet[v] = true

		if stats.IsNum {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				numValues = append(numValues, f)
				stats.Sum += f
				if first || v < minVal {
					minVal = v
					first = false
				}
				if first || v > maxVal {
					maxVal = v
					first = false
				}
			}
		} else {
			if first || v < minVal {
				minVal = v
				first = false
			}
			if first || v > maxVal {
				maxVal = v
				first = false
			}
		}
	}

	stats.Unique = len(uniqueSet)
	stats.Min = minVal
	stats.Max = maxVal

	if stats.Count > 0 && stats.IsNum {
		stats.Mean = stats.Sum / float64(stats.Count)

		if len(numValues) > 0 {
			// Median
			sorted := make([]float64, len(numValues))
			copy(sorted, numValues)
			sort.Float64s(sorted)
			mid := len(sorted) / 2
			if len(sorted)%2 == 0 {
				stats.Median = (sorted[mid-1] + sorted[mid]) / 2
			} else {
				stats.Median = sorted[mid]
			}

			// Standard Deviation
			var variance float64
			for _, v := range numValues {
				diff := v - stats.Mean
				variance += diff * diff
			}
			variance /= float64(len(numValues))
			stats.StdDev = math.Sqrt(variance)
		}
	}

	return stats
}

// StatsSummary returns stats for all columns.
func (t *Table) StatsSummary() []Stats {
	stats := make([]Stats, len(t.Headers))
	for i := range t.Headers {
		stats[i] = t.ColumnStats(i)
	}
	return stats
}

// RenderStats renders statistics as a formatted string.
func (t *Table) RenderStats() string {
	stats := t.StatsSummary()
	var sb strings.Builder

	for _, s := range stats {
		sb.WriteString(fmt.Sprintf("Column: %s\n", s.Column))
		sb.WriteString(fmt.Sprintf("  Type:   %s\n", typeStr(s.IsNum)))
		sb.WriteString(fmt.Sprintf("  Count:  %d\n", s.Count))
		sb.WriteString(fmt.Sprintf("  Unique: %d\n", s.Unique))
		sb.WriteString(fmt.Sprintf("  Empty:  %d\n", s.Empty))

		if s.IsNum {
			sb.WriteString(fmt.Sprintf("  Min:    %s\n", s.Min))
			sb.WriteString(fmt.Sprintf("  Max:    %s\n", s.Max))
			sb.WriteString(fmt.Sprintf("  Sum:    %.4f\n", s.Sum))
			sb.WriteString(fmt.Sprintf("  Mean:   %.4f\n", s.Mean))
			sb.WriteString(fmt.Sprintf("  Median: %.4f\n", s.Median))
			sb.WriteString(fmt.Sprintf("  StdDev: %.4f\n", s.StdDev))
		} else {
			sb.WriteString(fmt.Sprintf("  Min:    %s\n", s.Min))
			sb.WriteString(fmt.Sprintf("  Max:    %s\n", s.Max))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func typeStr(isNum bool) string {
	if isNum {
		return "numeric"
	}
	return "text"
}
