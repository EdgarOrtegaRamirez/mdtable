package table_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/mdtable/internal/table"
)

func TestParseJSONFormat(t *testing.T) {
	json := `[{"Name":"Alice","Score":"90"},{"Name":"Bob","Score":"75"}]`

	tbl, err := table.ParseJSON(json)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if tbl.NumCols() != 2 {
		t.Errorf("expected 2 columns, got %d: %v", tbl.NumCols(), tbl.Headers)
	}
	if tbl.NumRows() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.NumRows())
	}
}
