package export_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"pos/internal/domain"
	"pos/internal/export"
)

func TestCSVExportAndEscaping(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results.csv")
	name := "Анна, \"Иванова\"\nгруппа 1"
	err := (export.CSVExporter{}).Export(path, []domain.Result{
		{StudentID: "s1", StudentName: name, Score: 7.7, Graded: 2, Total: 2, Policy: "zero"},
	})
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0][0] != "student_id" || rows[1][1] != name || rows[1][2] != "7.70" || rows[1][5] != "zero" {
		t.Fatalf("неверный CSV: %#v", rows)
	}
}

func TestCSVExportReportsInvalidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "results.csv")
	if err := (export.CSVExporter{}).Export(path, nil); err == nil {
		t.Fatal("ошибка записи потеряна")
	}
}
