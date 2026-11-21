package app

import "fmt"

func (c *CLI) RunDemo(exportPath string) error {
	fmt.Fprintln(c.output, "ПОС. Демонстрация работы ведомости")
	fmt.Fprintln(c.output, "1. Добавляем студентов и критерии.")
	for _, student := range []struct{ id, name string }{
		{"s1", "Анна Иванова"}, {"s2", "Иван Петров"},
	} {
		if err := c.service.AddStudent(student.id, student.name); err != nil {
			return err
		}
	}
	if err := c.service.AddCriterion("hw", "Домашняя работа", 10, 0.4); err != nil {
		return err
	}
	if err := c.service.AddCriterion("exam", "Экзамен", 100, 0.6); err != nil {
		return err
	}
	if err := c.showStudents(); err != nil {
		return err
	}
	if err := c.showCriteria(); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "\n2. Выставляем оценки.")
	for _, evaluation := range []struct {
		studentID, criterionID string
		score                  float64
	}{
		{"s1", "hw", 8}, {"s1", "exam", 75}, {"s2", "hw", 6},
	} {
		if err := c.service.SetEvaluation(evaluation.studentID, evaluation.criterionID, evaluation.score, ""); err != nil {
			return err
		}
	}
	if err := c.showEvaluations(); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "\n3. Проверяем запрет некорректных оценок.")
	if err := c.service.SetEvaluation("s1", "hw", 12, ""); err != nil {
		fmt.Fprintln(c.output, "Ожидаемый отказ:", err)
	} else {
		return fmt.Errorf("домен ошибочно разрешил 12 баллов из 10")
	}
	if err := c.service.SetEvaluation("s1", "unknown", 5, ""); err != nil {
		fmt.Fprintln(c.output, "Ожидаемый отказ:", err)
	} else {
		return fmt.Errorf("домен ошибочно разрешил неизвестный критерий")
	}
	fmt.Fprintln(c.output, "\n4. Исправляем оценку s1 за hw с 8 на 9.")
	if err := c.service.SetEvaluation("s1", "hw", 9, "Исправлена после проверки"); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "Отменяем оценку s2 за hw, проверяем список и выставляем заново.")
	if err := c.service.CancelEvaluation("s2", "hw"); err != nil {
		return err
	}
	if err := c.showEvaluations(); err != nil {
		return err
	}
	if err := c.service.SetEvaluation("s2", "hw", 6, "Повторная проверка"); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "\n5. Считаем итог по текущим оценкам.")
	if err := c.showResults(); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "\n6. Экспортируем итог.")
	if err := c.service.ExportResults(exportPath); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "Итоги записаны в", exportPath)
	fmt.Fprintln(c.output, "Для работы с пустой ведомостью: go run ./cmd/pos --interactive")
	return nil
}
