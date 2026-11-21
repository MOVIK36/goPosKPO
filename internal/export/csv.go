package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"pos/internal/domain"
)

type CSVExporter struct{}

func (CSVExporter) Export(path string, results []domain.Result) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("создание CSV: %w", err)
	}
	writer := csv.NewWriter(file)
	rows := [][]string{{"student_id", "student_name", "score", "graded", "total", "policy"}}
	for _, result := range results {
		rows = append(rows, []string{
			result.StudentID, result.StudentName, strconv.FormatFloat(result.Score, 'f', 2, 64),
			strconv.Itoa(result.Graded), strconv.Itoa(result.Total), result.Policy,
		})
	}
	err = writer.WriteAll(rows)
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("запись CSV: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("закрытие CSV: %w", closeErr)
	}
	return nil
}
