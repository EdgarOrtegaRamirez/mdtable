package table

import (
	"testing"
)

func TestParseBasic(t *testing.T) {
	input := `| Name  | Age | City    |
|-------|-----|---------|
| Alice | 30  | NYC     |
| Bob   | 25  | London  |
| Carol | 35  | Paris   |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if tbl.NumCols() != 3 {
		t.Errorf("expected 3 columns, got %d", tbl.NumCols())
	}
	if tbl.NumRows() != 3 {
		t.Errorf("expected 3 rows, got %d", tbl.NumRows())
	}
	if tbl.Headers[0] != "Name" {
		t.Errorf("expected header 'Name', got %q", tbl.Headers[0])
	}
	if tbl.Rows[0][0] != "Alice" {
		t.Errorf("expected 'Alice', got %q", tbl.Rows[0][0])
	}
}

func TestParseAlignment(t *testing.T) {
	input := `| Left | Center | Right |
|------|:------:|------:|
| a    | b      | c     |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if tbl.Alignment[0] != AlignLeft {
		t.Errorf("expected AlignLeft, got %v", tbl.Alignment[0])
	}
	if tbl.Alignment[1] != AlignCenter {
		t.Errorf("expected AlignCenter, got %v", tbl.Alignment[1])
	}
	if tbl.Alignment[2] != AlignRight {
		t.Errorf("expected AlignRight, got %v", tbl.Alignment[2])
	}
}

func TestRender(t *testing.T) {
	tbl := New([]string{"A", "B"})
	tbl.AddRow([]string{"1", "2"})
	tbl.AddRow([]string{"10", "20"})

	output := tbl.Render()
	if output == "" {
		t.Error("Render returned empty string")
	}

	// Check it contains the headers
	if !contains(output, "| A") {
		t.Error("output missing header A")
	}
	if !contains(output, "| B") {
		t.Error("output missing header B")
	}
}

func TestSortByColumnNumeric(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tbl.SortByColumn(1, false) // Ascending

	if tbl.Rows[0][0] != "Bob" {
		t.Errorf("expected Bob first, got %q", tbl.Rows[0][0])
	}
	if tbl.Rows[2][0] != "Alice" {
		t.Errorf("expected Alice last, got %q", tbl.Rows[2][0])
	}
}

func TestSortByColumnDescending(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tbl.SortByColumn(1, true) // Descending

	if tbl.Rows[0][0] != "Alice" {
		t.Errorf("expected Alice first, got %q", tbl.Rows[0][0])
	}
	if tbl.Rows[2][0] != "Bob" {
		t.Errorf("expected Bob last, got %q", tbl.Rows[2][0])
	}
}

func TestFilterRegex(t *testing.T) {
	input := `| Name  | City    |
|-------|---------|
| Alice | NYC     |
| Bob   | London  |
| Carol | Paris   |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := tbl.Filter(1, "^N")
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if result.NumRows() != 1 {
		t.Errorf("expected 1 row, got %d", result.NumRows())
	}
	if result.Rows[0][0] != "Alice" {
		t.Errorf("expected Alice, got %q", result.Rows[0][0])
	}
}

func TestFilterEqual(t *testing.T) {
	input := `| Name  | City    |
|-------|---------|
| Alice | NYC     |
| Bob   | London  |
| Carol | Paris   |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.FilterEqual(1, "london")
	if result.NumRows() != 1 {
		t.Errorf("expected 1 row, got %d", result.NumRows())
	}
	if result.Rows[0][0] != "Bob" {
		t.Errorf("expected Bob, got %q", result.Rows[0][0])
	}
}

func TestFilterContains(t *testing.T) {
	input := `| Product | Price |
|---------|-------|
| Widget  | 10    |
| WidgetX | 20    |
| Gadget  | 15    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.FilterContains(0, "widget")
	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
}

func TestFilterGreaterThan(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.FilterGreaterThan(1, 80)
	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
}

func TestHead(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |
| Dave  | 60    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.Head(2)
	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
	if result.Rows[0][0] != "Alice" {
		t.Errorf("expected Alice, got %q", result.Rows[0][0])
	}
}

func TestTail(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |
| Dave  | 60    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.Tail(2)
	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
	if result.Rows[0][0] != "Carol" {
		t.Errorf("expected Carol, got %q", result.Rows[0][0])
	}
}

func TestUnique(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Alice | 90    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.Unique()
	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
}

func TestUniqueByColumn(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Alice | 85    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.UniqueByColumn(0)
	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
}

func TestRenderCSV(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	csv := tbl.RenderCSV()
	if csv == "" {
		t.Error("RenderCSV returned empty string")
	}
	if !contains(csv, "Name,Score") {
		t.Error("CSV missing header")
	}
}

func TestRenderJSON(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	json := tbl.RenderJSON()
	if json == "" {
		t.Error("RenderJSON returned empty string")
	}
	if !contains(json, `"Name"`) {
		t.Error("JSON missing Name key")
	}
}

func TestRenderHTML(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := tbl.RenderHTML()
	if html == "" {
		t.Error("RenderHTML returned empty string")
	}
	if !contains(html, "<table>") {
		t.Error("HTML missing <table> tag")
	}
	if !contains(html, "<th>Name</th>") {
		t.Error("HTML missing header")
	}
}

func TestColumnStats(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stats := tbl.ColumnStats(1)
	if !stats.IsNum {
		t.Error("expected numeric column")
	}
	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}
	if stats.Mean != 83.33333333333333 {
		t.Errorf("expected mean ~83.33, got %f", stats.Mean)
	}
	if stats.Min != "75" {
		t.Errorf("expected min 75, got %s", stats.Min)
	}
	if stats.Max != "90" {
		t.Errorf("expected max 90, got %s", stats.Max)
	}
}

func TestMultiSort(t *testing.T) {
	input := `| Name  | Dept  | Score |
|-------|-------|-------|
| Alice | Eng   | 90    |
| Bob   | Eng   | 85    |
| Carol | Sales | 90    |
| Dave  | Sales | 75    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tbl.SortMulti([]SortKey{
		{Column: 1, Descending: false}, // Dept ascending
		{Column: 2, Descending: true},  // Score descending
	})

	if tbl.Rows[0][0] != "Alice" {
		t.Errorf("expected Alice first, got %q", tbl.Rows[0][0])
	}
	if tbl.Rows[1][0] != "Bob" {
		t.Errorf("expected Bob second, got %q", tbl.Rows[1][0])
	}
}

func TestAddColumn(t *testing.T) {
	tbl := New([]string{"Name"})
	tbl.AddRow([]string{"Alice"})
	tbl.AddRow([]string{"Bob"})

	tbl.AddColumn("Score", []string{"90", "75"})

	if tbl.NumCols() != 2 {
		t.Errorf("expected 2 columns, got %d", tbl.NumCols())
	}
	if tbl.Rows[0][1] != "90" {
		t.Errorf("expected 90, got %q", tbl.Rows[0][1])
	}
}

func TestRemoveColumn(t *testing.T) {
	tbl := New([]string{"Name", "Score", "City"})
	tbl.AddRow([]string{"Alice", "90", "NYC"})
	tbl.AddRow([]string{"Bob", "75", "London"})

	tbl.RemoveColumn(1) // Remove Score

	if tbl.NumCols() != 2 {
		t.Errorf("expected 2 columns, got %d", tbl.NumCols())
	}
	if tbl.Headers[0] != "Name" {
		t.Errorf("expected Name, got %q", tbl.Headers[0])
	}
	if tbl.Headers[1] != "City" {
		t.Errorf("expected City, got %q", tbl.Headers[1])
	}
}

func TestClone(t *testing.T) {
	tbl := New([]string{"A", "B"})
	tbl.AddRow([]string{"1", "2"})

	clone := tbl.Clone()
	clone.Rows[0][0] = "99"

	if tbl.Rows[0][0] != "1" {
		t.Error("Clone modified original")
	}
}

func TestColumnIndex(t *testing.T) {
	tbl := New([]string{"Name", "Score", "City"})

	if idx := tbl.ColumnIndex("score"); idx != 1 {
		t.Errorf("expected 1, got %d", idx)
	}
	if idx := tbl.ColumnIndex("missing"); idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}

func TestParseCSVFormat(t *testing.T) {
	csv := `Name,Score,City
Alice,90,NYC
Bob,75,London`

	tbl, err := ParseCSV(csv)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if tbl.NumCols() != 3 {
		t.Errorf("expected 3 columns, got %d", tbl.NumCols())
	}
	if tbl.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.NumRows())
	}
	if tbl.Rows[0][0] != "Alice" {
		t.Errorf("expected Alice, got %q", tbl.Rows[0][0])
	}
}

func TestParseJSONFormat(t *testing.T) {
	json := `[{"Name":"Alice","Score":"90"},{"Name":"Bob","Score":"75"}]`

	tbl, err := ParseJSON(json)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if tbl.NumCols() != 2 {
		t.Errorf("expected 2 columns, got %d", tbl.NumCols())
	}
	if tbl.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.NumRows())
	}
}

func TestRenderTSV(t *testing.T) {
	tbl := New([]string{"A", "B"})
	tbl.AddRow([]string{"1", "2"})

	tsv := tbl.RenderTSV()
	if tsv == "" {
		t.Error("RenderTSV returned empty string")
	}
	if !contains(tsv, "A\tB") {
		t.Error("TSV missing tab-separated header")
	}
}

func TestDiffEqual(t *testing.T) {
	old := New([]string{"Name", "Score"})
	old.AddRow([]string{"Alice", "90"})
	old.AddRow([]string{"Bob", "75"})

	new := New([]string{"Name", "Score"})
	new.AddRow([]string{"Alice", "90"})
	new.AddRow([]string{"Bob", "75"})

	diffs, err := Diff(old, new)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	stats := ComputeDiffStats(diffs)
	if stats.Added != 0 || stats.Removed != 0 || stats.Modified != 0 {
		t.Errorf("expected no changes, got +%d -%d ~%d", stats.Added, stats.Removed, stats.Modified)
	}
}

func TestDiffAdded(t *testing.T) {
	old := New([]string{"Name", "Score"})
	old.AddRow([]string{"Alice", "90"})

	new := New([]string{"Name", "Score"})
	new.AddRow([]string{"Alice", "90"})
	new.AddRow([]string{"Bob", "75"})

	diffs, err := Diff(old, new)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	stats := ComputeDiffStats(diffs)
	if stats.Added != 1 {
		t.Errorf("expected 1 added, got %d", stats.Added)
	}
}

func TestDiffRemoved(t *testing.T) {
	old := New([]string{"Name", "Score"})
	old.AddRow([]string{"Alice", "90"})
	old.AddRow([]string{"Bob", "75"})

	new := New([]string{"Name", "Score"})
	new.AddRow([]string{"Alice", "90"})

	diffs, err := Diff(old, new)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	stats := ComputeDiffStats(diffs)
	if stats.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", stats.Removed)
	}
}

func TestDiffModified(t *testing.T) {
	old := New([]string{"Name", "Score"})
	old.AddRow([]string{"Alice", "90"})

	new := New([]string{"Name", "Score"})
	new.AddRow([]string{"Alice", "95"})

	diffs, err := Diff(old, new)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	stats := ComputeDiffStats(diffs)
	if stats.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", stats.Modified)
	}
}

func TestMergeInner(t *testing.T) {
	left := New([]string{"Name", "Score"})
	left.AddRow([]string{"Alice", "90"})
	left.AddRow([]string{"Bob", "75"})
	left.AddRow([]string{"Carol", "85"})

	right := New([]string{"Name", "City"})
	right.AddRow([]string{"Alice", "NYC"})
	right.AddRow([]string{"Bob", "London"})

	merged, err := Merge(left, right, "Name", "Name", MergeInner)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if merged.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", merged.NumRows())
	}
	if merged.NumCols() != 3 {
		t.Errorf("expected 3 columns, got %d", merged.NumCols())
	}
}

func TestMergeLeft(t *testing.T) {
	left := New([]string{"Name", "Score"})
	left.AddRow([]string{"Alice", "90"})
	left.AddRow([]string{"Bob", "75"})

	right := New([]string{"Name", "City"})
	right.AddRow([]string{"Alice", "NYC"})

	merged, err := Merge(left, right, "Name", "Name", MergeLeft)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if merged.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", merged.NumRows())
	}
}

func TestConcatenate(t *testing.T) {
	t1 := New([]string{"Name", "Score"})
	t1.AddRow([]string{"Alice", "90"})

	t2 := New([]string{"Name", "Score"})
	t2.AddRow([]string{"Bob", "75"})

	err := t1.Concatenate(t2)
	if err != nil {
		t.Fatalf("Concatenate failed: %v", err)
	}

	if t1.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", t1.NumRows())
	}
}

func TestSubset(t *testing.T) {
	tbl := New([]string{"A", "B", "C"})
	tbl.AddRow([]string{"1", "2", "3"})

	sub := tbl.Subset([]int{0, 2})

	if sub.NumCols() != 2 {
		t.Errorf("expected 2 columns, got %d", sub.NumCols())
	}
	if sub.Headers[0] != "A" || sub.Headers[1] != "C" {
		t.Errorf("unexpected headers: %v", sub.Headers)
	}
}

func TestIsNumericColumn(t *testing.T) {
	tbl := New([]string{"Num", "Text"})
	tbl.AddRow([]string{"1", "hello"})
	tbl.AddRow([]string{"2.5", "world"})
	tbl.AddRow([]string{"-3", "foo"})

	if !tbl.IsNumericColumn(0) {
		t.Error("expected column 0 to be numeric")
	}
	if tbl.IsNumericColumn(1) {
		t.Error("expected column 1 to not be numeric")
	}
}

func TestFormatError(t *testing.T) {
	_, err := Parse("too few lines")
	if err == nil {
		t.Error("expected error for malformed table")
	}
}

func TestRenderDiff(t *testing.T) {
	diffs := []DiffRow{
		{Op: DiffEqual, Row: []string{"Alice", "90"}},
		{Op: DiffAdded, Row: []string{"Bob", "75"}},
		{Op: DiffRemoved, Row: []string{"Carol", "85"}},
	}

	output := RenderDiff(diffs, nil)
	if output == "" {
		t.Error("RenderDiff returned empty string")
	}
	if !contains(output, "+ Bob") {
		t.Error("output missing added row")
	}
	if !contains(output, "- Carol") {
		t.Error("output missing removed row")
	}
}

func TestParseRowEdgeCases(t *testing.T) {
	// Table without leading/trailing pipes
	input := `Name | Score
-----|------
Alice| 90`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if tbl.NumCols() != 2 {
		t.Errorf("expected 2 columns, got %d", tbl.NumCols())
	}
}

func TestFilterNot(t *testing.T) {
	input := `| Name  | City    |
|-------|---------|
| Alice | NYC     |
| Bob   | London  |
| Carol | Paris   |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := tbl.FilterNot(1, "^N")
	if err != nil {
		t.Fatalf("FilterNot failed: %v", err)
	}

	if result.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.NumRows())
	}
}

func TestFilterLessThan(t *testing.T) {
	input := `| Name  | Score |
|-------|-------|
| Alice | 90    |
| Bob   | 75    |
| Carol | 85    |`

	tbl, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := tbl.FilterLessThan(1, 85)
	if result.NumRows() != 1 {
		t.Errorf("expected 1 row, got %d", result.NumRows())
	}
	if result.Rows[0][0] != "Bob" {
		t.Errorf("expected Bob, got %q", result.Rows[0][0])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
