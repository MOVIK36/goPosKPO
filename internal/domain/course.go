package domain

import (
	"fmt"
	"maps"
	"sort"
	"strings"
)

type evaluationKey struct {
	studentID   string
	criterionID string
}

type Course struct {
	students    map[string]Student
	criteria    map[string]Criterion
	evaluations map[evaluationKey]Evaluation
}

func NewCourse() *Course {
	c := &Course{}
	c.init()
	return c
}

func (c *Course) init() {
	if c.students == nil {
		c.students = make(map[string]Student)
	}
	if c.criteria == nil {
		c.criteria = make(map[string]Criterion)
	}
	if c.evaluations == nil {
		c.evaluations = make(map[evaluationKey]Evaluation)
	}
}

func (c *Course) AddStudent(id, name string) error {
	student, err := NewStudent(id, name)
	if err != nil {
		return err
	}
	if _, exists := c.students[student.ID()]; exists {
		return fmt.Errorf("студент с ID %q уже существует", student.ID())
	}
	c.init()
	c.students[student.ID()] = student
	return nil
}

func (c *Course) RenameStudent(id, name string) error {
	student, err := NewStudent(id, name)
	if err != nil {
		return err
	}
	if _, exists := c.students[student.ID()]; !exists {
		return fmt.Errorf("студент %q не найден", student.ID())
	}
	c.students[student.ID()] = student
	return nil
}

func (c *Course) RemoveStudent(id string) error {
	id = strings.TrimSpace(id)
	if _, exists := c.students[id]; !exists {
		return fmt.Errorf("студент %q не найден", id)
	}
	delete(c.students, id)
	for key := range c.evaluations {
		if key.studentID == id {
			delete(c.evaluations, key)
		}
	}
	return nil
}

func (c *Course) AddCriterion(id, name string, maxScore, weight float64) error {
	criterion, err := NewCriterion(id, name, maxScore, weight)
	if err != nil {
		return err
	}
	if _, exists := c.criteria[criterion.ID()]; exists {
		return fmt.Errorf("критерий с ID %q уже существует", criterion.ID())
	}
	c.init()
	c.criteria[criterion.ID()] = criterion
	return nil
}

func (c *Course) UpdateCriterion(id, name string, maxScore, weight float64) error {
	criterion, err := NewCriterion(id, name, maxScore, weight)
	if err != nil {
		return err
	}
	if _, exists := c.criteria[criterion.ID()]; !exists {
		return fmt.Errorf("критерий %q не найден", criterion.ID())
	}
	for _, evaluation := range c.evaluations {
		if evaluation.CriterionID() == criterion.ID() && evaluation.Score().Value() > maxScore {
			return fmt.Errorf("новый максимум меньше уже выставленной оценки %g", evaluation.Score().Value())
		}
	}
	c.criteria[criterion.ID()] = criterion
	return nil
}

func (c *Course) RemoveCriterion(id string) error {
	id = strings.TrimSpace(id)
	if _, exists := c.criteria[id]; !exists {
		return fmt.Errorf("критерий %q не найден", id)
	}
	delete(c.criteria, id)
	for key := range c.evaluations {
		if key.criterionID == id {
			delete(c.evaluations, key)
		}
	}
	return nil
}

func (c *Course) SetEvaluation(studentID, criterionID string, value float64, comment string) error {
	evaluation, err := (EvaluationFactory{}).Create(c, studentID, criterionID, value, comment)
	if err != nil {
		return err
	}
	c.init()
	key := evaluationKey{evaluation.StudentID(), evaluation.CriterionID()}
	c.evaluations[key] = evaluation
	return nil
}

func (c *Course) CancelEvaluation(studentID, criterionID string) error {
	key := evaluationKey{strings.TrimSpace(studentID), strings.TrimSpace(criterionID)}
	if _, exists := c.evaluations[key]; !exists {
		return fmt.Errorf("оценка студента %q по критерию %q не найдена", key.studentID, key.criterionID)
	}
	delete(c.evaluations, key)
	return nil
}

func (c *Course) Student(id string) (Student, bool) {
	student, exists := c.students[strings.TrimSpace(id)]
	return student, exists
}

func (c *Course) Students() []Student {
	students := make([]Student, 0, len(c.students))
	for _, student := range c.students {
		students = append(students, student)
	}
	sort.Slice(students, func(i, j int) bool { return students[i].ID() < students[j].ID() })
	return students
}

func (c *Course) Criteria() []Criterion {
	criteria := make([]Criterion, 0, len(c.criteria))
	for _, criterion := range c.criteria {
		criteria = append(criteria, criterion)
	}
	sort.Slice(criteria, func(i, j int) bool { return criteria[i].ID() < criteria[j].ID() })
	return criteria
}

func (c *Course) Evaluations() []Evaluation {
	evaluations := make([]Evaluation, 0, len(c.evaluations))
	for _, evaluation := range c.evaluations {
		evaluations = append(evaluations, evaluation)
	}
	sort.Slice(evaluations, func(i, j int) bool {
		if evaluations[i].StudentID() == evaluations[j].StudentID() {
			return evaluations[i].CriterionID() < evaluations[j].CriterionID()
		}
		return evaluations[i].StudentID() < evaluations[j].StudentID()
	})
	return evaluations
}

func (c *Course) Clone() *Course {
	return &Course{
		students:    maps.Clone(c.students),
		criteria:    maps.Clone(c.criteria),
		evaluations: maps.Clone(c.evaluations),
	}
}
