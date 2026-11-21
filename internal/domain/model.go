package domain

import (
	"fmt"
	"math"
	"strings"
)

type Student struct {
	id   string
	name string
}

func NewStudent(id, name string) (Student, error) {
	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || name == "" {
		return Student{}, fmt.Errorf("ID и имя студента не должны быть пустыми")
	}
	return Student{id: id, name: name}, nil
}

func (s Student) ID() string   { return s.id }
func (s Student) Name() string { return s.name }

type Criterion struct {
	id       string
	name     string
	maxScore float64
	weight   float64
}

func NewCriterion(id, name string, maxScore, weight float64) (Criterion, error) {
	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || name == "" {
		return Criterion{}, fmt.Errorf("ID и название критерия не должны быть пустыми")
	}
	if !finite(maxScore) || maxScore <= 0 {
		return Criterion{}, fmt.Errorf("максимальный балл должен быть положительным конечным числом")
	}
	if !finite(weight) || weight <= 0 {
		return Criterion{}, fmt.Errorf("вес должен быть положительным конечным числом")
	}
	return Criterion{id: id, name: name, maxScore: maxScore, weight: weight}, nil
}

func (c Criterion) ID() string        { return c.id }
func (c Criterion) Name() string      { return c.name }
func (c Criterion) MaxScore() float64 { return c.maxScore }
func (c Criterion) Weight() float64   { return c.weight }

type Score struct {
	value float64
}

func NewScore(value, maxScore float64) (Score, error) {
	if !finite(maxScore) || maxScore <= 0 || !finite(value) || value < 0 || value > maxScore {
		return Score{}, fmt.Errorf("балл должен быть конечным числом от 0 до %g", maxScore)
	}
	return Score{value: value}, nil
}

func (s Score) Value() float64 { return s.value }

type Evaluation struct {
	studentID   string
	criterionID string
	score       Score
	comment     string
}

func (e Evaluation) StudentID() string   { return e.studentID }
func (e Evaluation) CriterionID() string { return e.criterionID }
func (e Evaluation) Score() Score        { return e.score }
func (e Evaluation) Comment() string     { return e.comment }

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
