package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"pos/internal/application"
)

type CLI struct {
	service *application.Service
	input   *bufio.Scanner
	output  io.Writer
}

func NewCLI(service *application.Service, input io.Reader, output io.Writer) *CLI {
	return &CLI{service: service, input: bufio.NewScanner(input), output: output}
}

func (c *CLI) line(prompt string) (string, error) {
	fmt.Fprint(c.output, prompt)
	if !c.input.Scan() {
		if err := c.input.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(c.input.Text()), nil
}

func (c *CLI) number(prompt string) (float64, error) {
	text, err := c.line(prompt)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(text, ",", "."), 64)
	if err != nil {
		return 0, fmt.Errorf("введите число, например 8 или 7.5")
	}
	return value, nil
}

func (c *CLI) RunInteractive() error {
	fmt.Fprintln(c.output, "ПОС. Данные хранятся в памяти до выхода из программы.")
	for {
		fmt.Fprintln(c.output, `
1. Добавить студента
2. Список студентов
3. Добавить критерий
4. Рубрикатор
5. Выставить или исправить оценку
6. Список оценок
7. Отменить оценку
8. Итоги
9. Экспорт итогов в CSV
10. Изменить имя студента
11. Удалить студента и его оценки
12. Изменить критерий
13. Удалить критерий и оценки по нему
0. Выход`)
		choice, err := c.line("Выберите пункт: ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if choice == "0" {
			return nil
		}
		err = c.command(choice)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			fmt.Fprintln(c.output, "Ошибка:", err)
		}
	}
}

func (c *CLI) command(choice string) error {
	switch choice {
	case "1", "10":
		id, err := c.line("ID студента: ")
		if err != nil {
			return err
		}
		name, err := c.line("Имя студента: ")
		if err != nil {
			return err
		}
		if choice == "1" {
			err = c.service.AddStudent(id, name)
		} else {
			err = c.service.RenameStudent(id, name)
		}
		return c.done(err)
	case "2":
		return c.showStudents()
	case "3", "12":
		return c.editCriterion(choice == "12")
	case "4":
		return c.showCriteria()
	case "5", "7":
		return c.editEvaluation(choice == "7")
	case "6":
		return c.showEvaluations()
	case "8":
		return c.showResults()
	case "9":
		path, err := c.line("Путь к CSV (Enter - results.csv): ")
		if err != nil {
			return err
		}
		if path == "" {
			path = "results.csv"
		}
		if err := c.service.ExportResults(path); err != nil {
			return err
		}
		fmt.Fprintln(c.output, "Итоги записаны в", path)
		return nil
	case "11", "13":
		id, err := c.line("ID удаляемого объекта (связанные оценки тоже удалятся): ")
		if err != nil {
			return err
		}
		if choice == "11" {
			err = c.service.RemoveStudent(id)
		} else {
			err = c.service.RemoveCriterion(id)
		}
		return c.done(err)
	default:
		return fmt.Errorf("нет такого пункта меню")
	}
}

func (c *CLI) editCriterion(update bool) error {
	id, err := c.line("ID критерия: ")
	if err != nil {
		return err
	}
	name, err := c.line("Название: ")
	if err != nil {
		return err
	}
	maxScore, err := c.number("Максимальный балл: ")
	if err != nil {
		return err
	}
	weight, err := c.number("Вес (например 0.4): ")
	if err != nil {
		return err
	}
	if update {
		err = c.service.UpdateCriterion(id, name, maxScore, weight)
	} else {
		err = c.service.AddCriterion(id, name, maxScore, weight)
	}
	return c.done(err)
}

func (c *CLI) editEvaluation(cancel bool) error {
	studentID, err := c.line("ID студента: ")
	if err != nil {
		return err
	}
	criterionID, err := c.line("ID критерия: ")
	if err != nil {
		return err
	}
	if cancel {
		return c.done(c.service.CancelEvaluation(studentID, criterionID))
	}
	score, err := c.number("Балл: ")
	if err != nil {
		return err
	}
	comment, err := c.line("Комментарий (можно пустой): ")
	if err != nil {
		return err
	}
	return c.done(c.service.SetEvaluation(studentID, criterionID, score, comment))
}

func (c *CLI) done(err error) error {
	if err == nil {
		fmt.Fprintln(c.output, "Готово.")
	}
	return err
}

func (c *CLI) table() *tabwriter.Writer {
	return tabwriter.NewWriter(c.output, 0, 4, 2, ' ', 0)
}

func (c *CLI) showStudents() error {
	students, err := c.service.Students()
	if err != nil {
		return err
	}
	if len(students) == 0 {
		fmt.Fprintln(c.output, "Студентов пока нет.")
		return nil
	}
	w := c.table()
	fmt.Fprintln(w, "ID\tСтудент")
	for _, student := range students {
		fmt.Fprintf(w, "%s\t%s\n", student.ID(), student.Name())
	}
	return w.Flush()
}

func (c *CLI) showCriteria() error {
	criteria, err := c.service.Criteria()
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		fmt.Fprintln(c.output, "Рубрикатор пока пуст.")
		return nil
	}
	w := c.table()
	fmt.Fprintln(w, "ID\tКритерий\tМаксимум\tВес")
	for _, criterion := range criteria {
		fmt.Fprintf(w, "%s\t%s\t%g\t%g\n", criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight())
	}
	return w.Flush()
}

func (c *CLI) showEvaluations() error {
	evaluations, err := c.service.Evaluations()
	if err != nil {
		return err
	}
	if len(evaluations) == 0 {
		fmt.Fprintln(c.output, "Оценок пока нет.")
		return nil
	}
	w := c.table()
	fmt.Fprintln(w, "Студент (ID)\tКритерий (ID)\tБалл\tКомментарий")
	for _, evaluation := range evaluations {
		fmt.Fprintf(w, "%s\t%s\t%g\t%s\n", evaluation.StudentID(), evaluation.CriterionID(), evaluation.Score().Value(), evaluation.Comment())
	}
	return w.Flush()
}

func (c *CLI) showResults() error {
	results, err := c.service.Results()
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Fprintln(c.output, "Студентов пока нет.")
		return nil
	}
	fmt.Fprintln(c.output, "Итог по шкале 0-10; zero - пропуски как нули, graded - только выставленные оценки.")
	fmt.Fprintln(c.output, "Если выставлены не все оценки, результат предварительный.")
	w := c.table()
	fmt.Fprintln(w, "ID\tСтудент\tИтог\tВыставлено\tПравило")
	for _, result := range results {
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%d/%d\t%s\n", result.StudentID, result.StudentName, result.Score, result.Graded, result.Total, result.Policy)
	}
	return w.Flush()
}
