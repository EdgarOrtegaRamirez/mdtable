package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build binary
	dir, err := os.MkdirTemp("", "mdtable-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "mdtable")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/mdtable")
	cmd.Dir = filepath.Join(findModuleRoot())
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	os.Exit(m.Run())
}

func findModuleRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

const testTable = `| Name  | Score | City    |
|-------|-------|---------|
| Alice | 90    | NYC     |
| Bob   | 75    | London  |
| Carol | 85    | Paris   |`

func runCLI(args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Write test table to stdin
	cmd.Stdin = bytes.NewBufferString(testTable)

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestCLI_Sort(t *testing.T) {
	stdout, _, err := runCLI("sort", "Score")
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	if !containsStr(stdout, "Bob") {
		t.Error("expected Bob in output")
	}
	// Bob should be first (lowest score)
	lines := splitLines(stdout)
	if len(lines) > 2 && !containsStr(lines[2], "Bob") {
		t.Errorf("expected Bob first in data rows, got: %s", lines[2])
	}
}

func TestCLI_SortDesc(t *testing.T) {
	stdout, _, err := runCLI("sort", "Score", "-d")
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	lines := splitLines(stdout)
	if len(lines) > 2 && !containsStr(lines[2], "Alice") {
		t.Errorf("expected Alice first (descending), got: %s", lines[2])
	}
}

func TestCLI_Filter(t *testing.T) {
	stdout, _, err := runCLI("filter", "City", "London")
	if err != nil {
		t.Fatalf("filter failed: %v", err)
	}
	if !containsStr(stdout, "Bob") {
		t.Error("expected Bob in filtered output")
	}
	if containsStr(stdout, "Alice") {
		t.Error("Alice should not be in filtered output")
	}
}

func TestCLI_Head(t *testing.T) {
	stdout, _, err := runCLI("head", "2")
	if err != nil {
		t.Fatalf("head failed: %v", err)
	}
	if !containsStr(stdout, "Alice") {
		t.Error("expected Alice in output")
	}
	if !containsStr(stdout, "Bob") {
		t.Error("expected Bob in output")
	}
	if containsStr(stdout, "Carol") {
		t.Error("Carol should not be in head(2) output")
	}
}

func TestCLI_Tail(t *testing.T) {
	stdout, _, err := runCLI("tail", "2")
	if err != nil {
		t.Fatalf("tail failed: %v", err)
	}
	if !containsStr(stdout, "Bob") {
		t.Error("expected Bob in output")
	}
	if !containsStr(stdout, "Carol") {
		t.Error("expected Carol in output")
	}
	if containsStr(stdout, "Alice") {
		t.Error("Alice should not be in tail(2) output")
	}
}

func TestCLI_Stats(t *testing.T) {
	stdout, _, err := runCLI("stats", "-c", "Score")
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !containsStr(stdout, "numeric") {
		t.Error("expected numeric type")
	}
	if !containsStr(stdout, "Mean") {
		t.Error("expected Mean in stats output")
	}
}

func TestCLI_Convert_CSV(t *testing.T) {
	stdout, _, err := runCLI("convert", "csv")
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if !containsStr(stdout, "Name,Score,City") {
		t.Error("expected CSV header")
	}
}

func TestCLI_Convert_JSON(t *testing.T) {
	stdout, _, err := runCLI("convert", "json")
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if !containsStr(stdout, `"Name"`) {
		t.Error("expected JSON key")
	}
}

func TestCLI_Convert_HTML(t *testing.T) {
	stdout, _, err := runCLI("convert", "html")
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if !containsStr(stdout, "<table>") {
		t.Error("expected <table> tag")
	}
}

func TestCLI_Convert_TSV(t *testing.T) {
	stdout, _, err := runCLI("convert", "tsv")
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if !containsStr(stdout, "Name\tScore") {
		t.Error("expected tab-separated header")
	}
}

func TestCLI_Unique(t *testing.T) {
	stdout, _, err := runCLI("unique")
	if err != nil {
		t.Fatalf("unique failed: %v", err)
	}
	if !containsStr(stdout, "Alice") {
		t.Error("expected Alice in output")
	}
}

func TestCLI_Columns(t *testing.T) {
	stdout, _, err := runCLI("columns")
	if err != nil {
		t.Fatalf("columns failed: %v", err)
	}
	if !containsStr(stdout, "Name") {
		t.Error("expected Name column")
	}
	if !containsStr(stdout, "numeric") || !containsStr(stdout, "text") {
		t.Error("expected column types")
	}
}

func TestCLI_Version(t *testing.T) {
	stdout, _, err := runCLI("--version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !containsStr(stdout, "mdtable") {
		t.Error("expected mdtable in version output")
	}
}

func TestCLI_Diff(t *testing.T) {
	// Create a temp file with a different table
	tmpFile, err := os.CreateTemp("", "mdtable-diff-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(`| Name  | Score | City    |
|-------|-------|---------|
| Alice | 95    | NYC     |
| Bob   | 75    | London  |
| Dave  | 80    | Berlin  |`)
	tmpFile.Close()

	stdout, stderr, err := runCLI("diff", tmpFile.Name())
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if !containsStr(stdout, "+") || !containsStr(stdout, "-") {
		t.Error("expected diff markers in output")
	}
	if !containsStr(stderr, "Summary") {
		t.Error("expected summary in stderr")
	}
}

func TestCLI_FilterNot(t *testing.T) {
	cmd := exec.Command(binaryPath, "filter", "City", "London", "--not")
	cmd.Stdin = bytes.NewBufferString(testTable)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("filter --not failed: %v", err)
	}
	if containsStr(stdout.String(), "Bob") {
		t.Error("Bob should be excluded with --not")
	}
	if !containsStr(stdout.String(), "Alice") {
		t.Error("Alice should be in output")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
