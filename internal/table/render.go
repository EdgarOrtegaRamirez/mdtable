package table

import (
	"fmt"
	"strings"
)

// Render writes the table as a markdown string.
func (t *Table) Render() string {
	if t.NumCols() == 0 {
		return ""
	}

 widths := t.columnWidths()

	var sb strings.Builder

	// Header row
	sb.WriteString(renderRow(t.Headers, widths, t.Alignment))
	sb.WriteString("\n")

	// Separator row
	sepCells := make([]string, len(widths))
	for i, w := range widths {
		sep := t.Alignment[i].String()
		if w > len(sep) {
			sep = strings.Repeat("-", w-len(sep)+1) + sep
		} else {
			sep = t.Alignment[i].String()
		}
		sepCells[i] = sep
	}
	sb.WriteString(renderRow(sepCells, widths, nil))
	sb.WriteString("\n")

	// Data rows
	for _, row := range t.Rows {
		sb.WriteString(renderRow(row, widths, nil))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderRow renders a single row with proper alignment.
func renderRow(cells []string, widths []int, alignment []Alignment) string {
	var sb strings.Builder
	sb.WriteString("| ")

	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}

		aligned := cell
		if alignment != nil && i < len(alignment) {
			aligned = alignText(cell, w, alignment[i])
		} else {
			aligned = alignText(cell, w, AlignLeft)
		}

		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString(aligned)
	}

	sb.WriteString(" |")
	return sb.String()
}

// alignText aligns text within a given width.
func alignText(text string, width int, align Alignment) string {
	if width <= 0 {
		return text
	}
	padding := width - len(text)
	if padding <= 0 {
		return text
	}

	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + text
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	default: // AlignLeft
		return text + strings.Repeat(" ", padding)
	}
}

// columnWidths computes the maximum width for each column.
func (t *Table) columnWidths() []int {
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	return widths
}

// RenderCSV converts the table to CSV format.
func (t *Table) RenderCSV() string {
	var sb strings.Builder

	// Header
	sb.WriteString(joinCSV(t.Headers))
	sb.WriteString("\n")

	// Rows
	for _, row := range t.Rows {
		sb.WriteString(joinCSV(row))
		sb.WriteString("\n")
	}

	return sb.String()
}

// joinCSV joins cells with commas, quoting fields that contain commas or quotes.
func joinCSV(cells []string) string {
	var parts []string
	for _, cell := range cells {
		if strings.ContainsAny(cell, ",\"") || strings.Contains(cell, "\n") {
			escaped := strings.ReplaceAll(cell, `"`, `""`)
			parts = append(parts, `"`+escaped+`"`)
		} else {
			parts = append(parts, cell)
		}
	}
	return strings.Join(parts, ",")
}

// RenderTSV converts the table to TSV format.
func (t *Table) RenderTSV() string {
	var sb strings.Builder

	// Header
	sb.WriteString(strings.Join(t.Headers, "\t"))
	sb.WriteString("\n")

	// Rows
	for _, row := range t.Rows {
		sb.WriteString(strings.Join(row, "\t"))
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderHTML converts the table to an HTML table string.
func (t *Table) RenderHTML() string {
	var sb strings.Builder
	sb.WriteString("<table>\n")

	// Header
	sb.WriteString("  <thead>\n    <tr>\n")
	for _, h := range t.Headers {
		sb.WriteString(fmt.Sprintf("      <th>%s</th>\n", escapeHTML(h)))
	}
	sb.WriteString("    </tr>\n  </thead>\n")

	// Body
	sb.WriteString("  <tbody>\n")
	for _, row := range t.Rows {
		sb.WriteString("    <tr>\n")
		for i, cell := range row {
			tag := "td"
			if i < len(t.Alignment) {
				switch t.Alignment[i] {
				case AlignCenter:
					sb.WriteString(fmt.Sprintf("      <%s style=\"text-align:center\">%s</%s>\n", tag, escapeHTML(cell), tag))
					continue
				case AlignRight:
					sb.WriteString(fmt.Sprintf("      <%s style=\"text-align:right\">%s</%s>\n", tag, escapeHTML(cell), tag))
					continue
				}
			}
			sb.WriteString(fmt.Sprintf("      <%s>%s</%s>\n", tag, escapeHTML(cell), tag))
		}
		sb.WriteString("    </tr>\n")
	}
	sb.WriteString("  </tbody>\n</table>\n")

	return sb.String()
}

// RenderJSON converts the table to a JSON array of objects.
func (t *Table) RenderJSON() string {
	var sb strings.Builder
	sb.WriteString("[\n")

	for i, row := range t.Rows {
		sb.WriteString("  {\n")
		for j, h := range t.Headers {
			cell := ""
			if j < len(row) {
				cell = row[j]
			}
			sb.WriteString(fmt.Sprintf("    %q: %q", h, cell))
			if j < len(t.Headers)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("  }")
		if i < len(t.Rows)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("]\n")
	return sb.String()
}

// RenderMarkdown renders as a markdown table (same as Render).
func (t *Table) RenderMarkdown() string {
	return t.Render()
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
