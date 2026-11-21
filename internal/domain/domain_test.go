package domain_test

import (
	"math"
	"testing"

	"pos/internal/domain"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func sampleCourse(t *testing.T) *domain.Course {
	t.Helper()
	c := domain.NewCourse()
	must(t, c.AddStudent("s1", "Анна Иванова"))
	must(t, c.AddCriterion("hw", "Домашняя работа", 10, 0.4))
	must(t, c.AddCriterion("exam", "Экзамен", 100, 0.6))
	return c
}

func TestScoreBounds(t *testing.T) {
	for _, value := range []float64{0, 5.5, 10} {
		score, err := domain.NewScore(value, 10)
		must(t, err)
		if score.Value() != value {
			t.Fatalf("балл: %g, ожидался %g", score.Value(), value)
		}
	}
	for _, value := range []float64{-0.1, 10.1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := domain.NewScore(value, 10); err == nil {
			t.Errorf("принят недопустимый балл %g", value)
		}
	}
	for _, maxScore := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err := domain.NewScore(0, maxScore); err == nil {
			t.Errorf("принят недопустимый максимум %g", maxScore)
		}
	}
}

func TestStudentAndCriterionValidation(t *testing.T) {
	if _, err := domain.NewStudent(" ", "Имя"); err == nil {
		t.Fatal("принят пустой ID")
	}
	if _, err := domain.NewStudent("s1", " "); err == nil {
		t.Fatal("принято пустое имя")
	}
	student, err := domain.NewStudent(" s1 ", " Анна Иванова ")
	must(t, err)
	if student.ID() != "s1" || student.Name() != "Анна Иванова" {
		t.Fatal("пробелы по краям не удалены")
	}
	for _, tc := range []struct {
		id, name    string
		max, weight float64
	}{
		{"", "Работа", 10, 1}, {"hw", " ", 10, 1},
		{"hw", "Работа", 0, 1}, {"hw", "Работа", -1, 1},
		{"hw", "Работа", math.Inf(1), 1}, {"hw", "Работа", math.NaN(), 1},
		{"hw", "Работа", 10, 0}, {"hw", "Работа", 10, -1},
		{"hw", "Работа", 10, math.NaN()}, {"hw", "Работа", 10, math.Inf(1)},
	} {
		if _, err := domain.NewCriterion(tc.id, tc.name, tc.max, tc.weight); err == nil {
			t.Errorf("принят некорректный критерий: %+v", tc)
		}
	}
}

func TestEvaluationInvariantsAndReplacement(t *testing.T) {
	c := sampleCourse(t)
	must(t, c.SetEvaluation("s1", "hw", 8, "Первая проверка"))
	for _, tc := range []struct {
		studentID, criterionID string
		score                  float64
	}{
		{"missing", "hw", 5}, {"s1", "missing", 5}, {"s1", "hw", 12},
		{"s1", "hw", -1}, {"s1", "hw", math.NaN()},
	} {
		if err := c.SetEvaluation(tc.studentID, tc.criterionID, tc.score, ""); err == nil {
			t.Errorf("принята некорректная оценка: %+v", tc)
		}
	}
	if got := c.Evaluations(); len(got) != 1 || got[0].Score().Value() != 8 {
		t.Fatal("неудачная операция изменила предыдущую оценку")
	}
	must(t, c.SetEvaluation("s1", "hw", 9, " Исправлено "))
	got := c.Evaluations()
	if len(got) != 1 || got[0].Score().Value() != 9 || got[0].Comment() != "Исправлено" {
		t.Fatal("исправление должно заменять оценку, а не добавлять вторую")
	}
	must(t, c.CancelEvaluation("s1", "hw"))
	if len(c.Evaluations()) != 0 {
		t.Fatal("отменённая оценка осталась в ведомости")
	}
	if err := c.CancelEvaluation("s1", "hw"); err == nil {
		t.Fatal("повторная отмена должна возвращать ошибку")
	}
	if _, err := (domain.EvaluationFactory{}).Create(nil, "s1", "hw", 1, ""); err == nil {
		t.Fatal("фабрика приняла nil вместо ведомости")
	}
}

func TestCourseChangesPreserveInvariants(t *testing.T) {
	c := sampleCourse(t)
	if err := c.AddStudent("s1", "Другое имя"); err == nil {
		t.Fatal("дубликат студента разрешён")
	}
	if err := c.AddCriterion("hw", "Дубль", 1, 1); err == nil {
		t.Fatal("дубликат критерия разрешён")
	}
	must(t, c.SetEvaluation("s1", "hw", 8, ""))
	if err := c.UpdateCriterion("hw", "Работа", 7, 1); err == nil {
		t.Fatal("максимум уменьшился ниже выставленной оценки")
	}
	for _, criterion := range c.Criteria() {
		if criterion.ID() == "hw" && criterion.MaxScore() != 10 {
			t.Fatal("неудачное изменение затронуло критерий")
		}
	}
	must(t, c.UpdateCriterion("hw", "Практика", 20, 2))
	must(t, c.RenameStudent("s1", "Анна Петрова"))
	student, _ := c.Student("s1")
	if student.Name() != "Анна Петрова" || len(c.Evaluations()) != 1 {
		t.Fatal("переименование потеряло студента или оценку")
	}
	must(t, c.RemoveCriterion("hw"))
	if len(c.Evaluations()) != 0 || len(c.Criteria()) != 1 {
		t.Fatal("удаление критерия оставило связанные оценки")
	}
	must(t, c.SetEvaluation("s1", "exam", 75, ""))
	must(t, c.RemoveStudent("s1"))
	if len(c.Students()) != 0 || len(c.Evaluations()) != 0 {
		t.Fatal("удаление студента оставило связанные оценки")
	}
}

func TestUnknownObjectsCannotBeChanged(t *testing.T) {
	c := sampleCourse(t)
	for _, operation := range []func() error{
		func() error { return c.RenameStudent("missing", "Имя") },
		func() error { return c.RemoveStudent("missing") },
		func() error { return c.UpdateCriterion("missing", "Работа", 10, 1) },
		func() error { return c.RemoveCriterion("missing") },
		func() error { return c.RenameStudent("s1", " ") },
		func() error { return c.UpdateCriterion("hw", "Работа", 0, 1) },
	} {
		if err := operation(); err == nil {
			t.Fatal("ожидалась ошибка")
		}
	}
}

func TestCourseDoesNotExposeItsCollections(t *testing.T) {
	c := sampleCourse(t)
	students := c.Students()
	students[0] = domain.Student{}
	criteria := c.Criteria()
	criteria[0] = domain.Criterion{}
	copy := c.Clone()
	must(t, copy.RemoveStudent("s1"))
	if student, exists := c.Student("s1"); !exists || student.Name() != "Анна Иванова" {
		t.Fatal("изменение копии затронуло исходный агрегат")
	}
	if len(c.Criteria()) != 2 {
		t.Fatal("исходный рубрикатор изменился")
	}
}

func TestFinalScorePolicies(t *testing.T) {
	for _, tc := range []struct {
		name     string
		policy   domain.ScorePolicy
		withExam bool
		want     float64
	}{
		{"полный zero", domain.MissingAsZero{}, true, 7.7},
		{"полный graded", domain.GradedOnly{}, true, 7.7},
		{"пропуск zero", domain.MissingAsZero{}, false, 3.2},
		{"пропуск graded", domain.GradedOnly{}, false, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := sampleCourse(t)
			must(t, c.SetEvaluation("s1", "hw", 8, ""))
			if tc.withExam {
				must(t, c.SetEvaluation("s1", "exam", 75, ""))
			}
			result, err := domain.NewFinalScoreService(tc.policy).Calculate(c, "s1")
			must(t, err)
			if math.Abs(result.Score-tc.want) > 1e-10 || result.Policy != tc.policy.Name() || result.Total != 2 {
				t.Fatalf("получен %+v, ожидался балл %g", result, tc.want)
			}
		})
	}
}

func TestFinalScoreIsRecalculated(t *testing.T) {
	c := sampleCourse(t)
	scorer := domain.NewFinalScoreService(domain.MissingAsZero{})
	must(t, c.SetEvaluation("s1", "hw", 8, ""))
	must(t, c.SetEvaluation("s1", "exam", 75, ""))
	must(t, c.SetEvaluation("s1", "hw", 9, ""))
	result, err := scorer.Calculate(c, "s1")
	must(t, err)
	if math.Abs(result.Score-8.1) > 1e-10 {
		t.Fatalf("исправление не учтено: %+v", result)
	}
	must(t, c.CancelEvaluation("s1", "exam"))
	result, err = scorer.Calculate(c, "s1")
	must(t, err)
	if math.Abs(result.Score-3.6) > 1e-10 || result.Graded != 1 {
		t.Fatalf("отмена не учтена: %+v", result)
	}
	must(t, c.UpdateCriterion("hw", "Работа", 20, 0.6))
	result, err = scorer.Calculate(c, "s1")
	must(t, err)
	if math.Abs(result.Score-2.25) > 1e-10 {
		t.Fatalf("изменение рубрикатора не учтено: %+v", result)
	}
}

func TestEmptyAndExtremeScores(t *testing.T) {
	c := sampleCourse(t)
	for _, policy := range []domain.ScorePolicy{domain.MissingAsZero{}, domain.GradedOnly{}} {
		scorer := domain.NewFinalScoreService(policy)
		result, err := scorer.Calculate(c, "s1")
		must(t, err)
		if result.Score != 0 || result.Graded != 0 {
			t.Fatalf("без оценок ожидался ноль: %+v", result)
		}
		if _, err := scorer.Calculate(c, "missing"); err == nil {
			t.Fatal("рассчитан итог неизвестного студента")
		}
		if _, err := scorer.Calculate(nil, "s1"); err == nil {
			t.Fatal("принята nil-ведомость")
		}
	}
	empty := domain.NewCourse()
	must(t, empty.AddStudent("s1", "Имя"))
	if _, err := domain.NewFinalScoreService(domain.MissingAsZero{}).Calculate(empty, "s1"); err == nil {
		t.Fatal("без рубрикатора ожидалась ошибка")
	}
	must(t, c.UpdateCriterion("hw", "Работа", 10, math.MaxFloat64))
	must(t, c.UpdateCriterion("exam", "Экзамен", 100, math.MaxFloat64))
	must(t, c.SetEvaluation("s1", "hw", 10, ""))
	must(t, c.SetEvaluation("s1", "exam", 100, ""))
	result, err := domain.NewFinalScoreService(domain.MissingAsZero{}).Calculate(c, "s1")
	must(t, err)
	if result.Score != 10 {
		t.Fatalf("переполнение при больших весах: %g", result.Score)
	}
}

func TestGradedOnlyCountsExplicitZero(t *testing.T) {
	c := sampleCourse(t)
	must(t, c.AddStudent("s2", "Иван"))
	must(t, c.SetEvaluation("s2", "exam", 100, ""))
	must(t, c.SetEvaluation("s1", "hw", 8, ""))

	scorer := domain.NewFinalScoreService(domain.GradedOnly{})
	before, err := scorer.Calculate(c, "s1")
	must(t, err)
	if before.Score != 8 || before.Graded != 1 {
		t.Fatalf("при пропуске экзамена должен учитываться только балл за ДЗ: %+v", before)
	}

	must(t, c.SetEvaluation("s1", "exam", 0, ""))
	after, err := scorer.Calculate(c, "s1")
	must(t, err)
	if math.Abs(after.Score-3.2) > 1e-10 || after.Graded != 2 {
		t.Fatalf("выставленный ноль должен учитываться с весом экзамена: %+v", after)
	}
}
