package table

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Parse reads a markdown table from a string.
func Parse(input string) (*Table, error) {
	return ParseReader(strings.NewReader(input))
}

// ParseReader reads a markdown table from an io.Reader.
func ParseReader(r io.Reader) (*Table, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return ParseLines(lines)
}

// ParseLines parses a markdown table from a slice of lines.
func ParseLines(lines []string) (*Table, error) {
	// Filter empty lines and trim
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		filtered = append(filtered, line)
	}

	if len(filtered) < 2 {
		return nil, &FormatError{Line: 0, Message: "table requires at least a header and separator row"}
	}

	// Parse header row
	headerLine := filtered[0]
	headers, err := parseRow(headerLine)
	if err != nil {
		return nil, &FormatError{Line: 1, Message: fmt.Sprintf("parsing header: %v", err)}
	}

	if len(headers) == 0 {
		return nil, &FormatError{Line: 1, Message: "no columns found in header row"}
	}

	// Parse separator row
	separatorLine := filtered[1]
	separatorCells, err := parseRow(separatorLine)
	if err != nil {
		return nil, &FormatError{Line: 2, Message: fmt.Sprintf("parsing separator: %v", err)}
	}

	// Validate separator
	alignment := make([]Alignment, len(headers))
	for i := range alignment {
		alignment[i] = AlignLeft // default
		if i < len(separatorCells) {
			sep := strings.TrimSpace(separatorCells[i])
			if isSeparator(sep) {
				alignment[i] = ParseAlignment(sep)
			}
		}
	}

	// Parse data rows
	var rows [][]string
	for i, line := range filtered[2:] {
		cells, err := parseRow(line)
		if err != nil {
			return nil, &FormatError{Line: i + 3, Message: fmt.Sprintf("parsing row: %v", err)}
		}

		// Pad or truncate to match header count
		row := make([]string, len(headers))
		for j := range row {
			if j < len(cells) {
				row[j] = strings.TrimSpace(cells[j])
			}
		}
		rows = append(rows, row)
	}

	return &Table{
		Headers:   headers,
		Alignment: alignment,
		Rows:      rows,
	}, nil
}

// parseRow splits a markdown table row into cells.
func parseRow(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	// Remove leading and trailing pipes
	if line[0] == '|' {
		line = line[1:]
	}
	if len(line) > 0 && line[len(line)-1] == '|' {
		line = line[:len(line)-1]
	}

	// Split by pipe
	parts := strings.Split(line, "|")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}

	return result, nil
}

// isSeparator checks if a cell is a valid separator.
func isSeparator(s string) bool {
	if len(s) < 3 {
		return false
	}
	s = strings.TrimSpace(s)
	hasColonLeft := s[0] == ':'
	hasColonRight := s[len(s)-1] == ':'

	inner := s
	if hasColonLeft {
		inner = inner[1:]
	}
	if hasColonRight && len(inner) > 0 {
		inner = inner[:len(inner)-1]
	}

	for _, c := range inner {
		if c != '-' {
			return false
		}
	}
	return len(inner) > 0
}

// ParseCSV parses a CSV string into a Table.
func ParseCSV(input string) (*Table, error) {
	return ParseCSVReader(strings.NewReader(input))
}

// ParseCSVReader reads CSV from an io.Reader into a Table.
func ParseCSVReader(r io.Reader) (*Table, error) {
	scanner := bufio.NewScanner(r)
	var lines [][]string

	for scanner.Scan() {
		line := scanner.Text()
		cells := parseCSVLine(line)
		lines = append(lines, cells)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	if len(lines) < 1 {
		return nil, fmt.Errorf("CSV must have at least a header row")
	}

	t := New(lines[0])
	for _, row := range lines[1:] {
		t.AddRow(row)
	}
	return t, nil
}

// parseCSVLine parses a single CSV line.
func parseCSVLine(line string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if inQuotes {
			if ch == '"' {
				if i+1 < len(line) && line[i+1] == '"' {
					current.WriteByte('"')
					i++
				} else {
					inQuotes = false
				}
			} else {
				current.WriteByte(ch)
			}
		} else {
			if ch == '"' {
				inQuotes = true
			} else if ch == ',' {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		}
	}
	result = append(result, strings.TrimSpace(current.String()))
	return result
}

// ParseJSON parses a JSON array of objects into a Table.
func ParseJSON(input string) (*Table, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "[") {
		return nil, fmt.Errorf("expected JSON array")
	}

	var headers []string
	var rows [][]string
	headerSet := make(map[string]bool)

	objects := splitJSONObjects(input)
	for _, obj := range objects {
		kvPairs := parseJSONObject(obj)
		row := make([]string, len(headers))

		for key, val := range kvPairs {
			if !headerSet[key] {
				headers = append(headers, key)
				headerSet[key] = true
				for i := range rows {
					rows[i] = append(rows[i], "")
				}
			}
			idx := indexOf(headers, key)
			for len(row) <= idx {
				row = append(row, "")
			}
			row[idx] = val
		}

		for len(row) < len(headers) {
			row = append(row, "")
		}
		rows = append(rows, row)
	}

	t := New(headers)
	t.Rows = rows
	return t, nil
}

// splitJSONObjects splits a JSON array into individual object strings.
func splitJSONObjects(input string) []string {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return nil
	}
	input = input[1:]
	if len(input) > 0 && input[len(input)-1] == ']' {
		input = input[:len(input)-1]
	}

	var objects []string
	depth := 0
	inString := false
	escaped := false
	start := 0

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				objects = append(objects, input[start:i+1])
			}
		}
	}

	return objects
}

// parseJSONObject parses a JSON object string into key-value pairs.
func parseJSONObject(obj string) map[string]string {
	result := make(map[string]string)
	obj = strings.TrimSpace(obj)
	if len(obj) < 2 {
		return result
	}
	obj = obj[1 : len(obj)-1] // remove { }

	state := 0 // 0=seeking_key, 1=in_key, 2=seeking_value, 3=in_value_string
	inString := false
	escaped := false
	var currentKey strings.Builder
	var currentValue strings.Builder

	for i := 0; i < len(obj); i++ {
		ch := obj[i]

		if escaped {
			escaped = false
			if state == 1 {
				currentKey.WriteByte(ch)
			} else if state == 3 {
				currentValue.WriteByte(ch)
			}
			continue
		}

		if ch == '\\' && inString {
			escaped = true
			continue
		}

		if ch == '"' {
			if inString {
				inString = false
				if state == 1 {
					state = 2 // done with key, seek value
				} else if state == 3 {
					state = 0 // done with value, seek next key or end
				}
			} else {
				inString = true
				if state == 0 || state == 2 {
					// Starting a string
					if state == 0 {
						state = 1 // starting a key
						currentKey.Reset()
					} else {
						state = 3 // starting a value
						currentValue.Reset()
					}
				}
			}
			continue
		}

		if inString {
			if state == 1 {
				currentKey.WriteByte(ch)
			} else if state == 3 {
				currentValue.WriteByte(ch)
			}
			continue
		}

		// Not in string
		switch ch {
		case ':':
			// Already handled by state transitions on closing quote
		case ',':
			// Store the key-value pair
			if currentKey.Len() > 0 || currentValue.Len() > 0 {
				result[currentKey.String()] = currentValue.String()
				currentKey.Reset()
				currentValue.Reset()
				state = 0
			}
		case ' ', '\t', '\n', '\r':
			// Skip whitespace
		}
	}

	// Store last pair
	if currentKey.Len() > 0 || currentValue.Len() > 0 {
		result[currentKey.String()] = currentValue.String()
	}

	return result
}

func indexOf(arr []string, val string) int {
	for i, s := range arr {
		if s == val {
			return i
		}
	}
	return -1
}
