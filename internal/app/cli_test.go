package app_test

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pos/internal/app"
	"pos/internal/di"
)

func TestDemo(t *testing.T) {
	for _, policy := range []string{"zero", "graded"} {
		t.Run(policy, func(t *testing.T) {
			container, err := di.NewDefaultContainer(policy)
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			path := filepath.Join(t.TempDir(), "demo.csv")
			cli := app.NewCLI(container.Service, strings.NewReader(""), &output)
			if err := cli.RunDemo(path); err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{"Ожидаемый отказ:", "8.10", "Итоги записаны в"} {
				if !strings.Contains(output.String(), text) {
					t.Errorf("нет ожидаемого вывода %q", text)
				}
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
			wantSecond := "2.40"
			if policy == "graded" {
				wantSecond = "6.00"
			}
			if len(rows) != 3 || rows[1][2] != "8.10" || rows[2][2] != wantSecond || rows[2][5] != policy {
				t.Fatalf("неверный итог демо: %#v", rows)
			}
		})
	}
}

func TestInteractiveScenario(t *testing.T) {
	container, err := di.NewDefaultContainer("zero")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "итоги.csv")
	input := strings.Join([]string{
		"2", "4", "6", "8",
		"1", "s1", "Анна Иванова", "1", "s2", "Иван",
		"10", "s2", "Иван Петров", "2",
		"3", "hw", "Работа", "10", "0,4",
		"3", "exam", "Экзамен", "100", "0.6", "4",
		"5", "s1", "hw", "8", "Проверено",
		"5", "s1", "exam", "75", "",
		"5", "s2", "hw", "6", "",
		"5", "s1", "hw", "12", "Ошибка",
		"5", "s1", "missing", "5", "Ошибка",
		"5", "s1", "hw", "9", "Исправлено", "6", "8",
		"7", "s2", "hw",
		"12", "hw", "Практика", "20", "0.4",
		"5", "s2", "hw", "10", "", "8",
		"9", path, "13", "exam", "11", "s2", "0", "",
	}, "\n")
	var output bytes.Buffer
	if err := app.NewCLI(container.Service, strings.NewReader(input), &output).RunInteractive(); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Студентов пока нет", "Рубрикатор пока пуст", "Оценок пока нет", "от 0 до 10", "отсутствует в рубрикаторе", "8.10", "6.30", "Итоги записаны в"} {
		if !strings.Contains(output.String(), text) {
			t.Errorf("нет ожидаемого сообщения %q", text)
		}
	}
	results, err := container.Service.Results()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Score != 4.5 || results[0].Total != 1 {
		t.Fatalf("удаление студента/критерия не отразилось в итоге: %+v", results)
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
	if len(rows) != 3 || rows[1][2] != "6.30" || rows[2][2] != "2.00" {
		t.Fatalf("CSV не соответствует состоянию на момент экспорта: %#v", rows)
	}
}

func TestInvalidInputAndEOF(t *testing.T) {
	for _, input := range []string{"", "1\ns1\n", "99\n3\nhw\nРабота\nabc\n0\n"} {
		container, err := di.NewDefaultContainer("zero")
		if err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if err := app.NewCLI(container.Service, strings.NewReader(input), &output).RunInteractive(); err != nil {
			t.Fatal(err)
		}
		students, err := container.Service.Students()
		if err != nil || len(students) != 0 {
			t.Fatal("незавершённый ввод изменил ведомость")
		}
		if strings.HasPrefix(input, "99") && !strings.Contains(output.String(), "введите число") {
			t.Fatal("неверный числовой ввод не обработан")
		}
	}
}
