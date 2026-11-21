package domain

import (
	"fmt"
	"strings"
)

type EvaluationFactory struct{}

func (EvaluationFactory) Create(course *Course, studentID, criterionID string, value float64, comment string) (Evaluation, error) {
	if course == nil {
		return Evaluation{}, fmt.Errorf("ведомость не задана")
	}
	studentID, criterionID = strings.TrimSpace(studentID), strings.TrimSpace(criterionID)
	if _, exists := course.students[studentID]; !exists {
		return Evaluation{}, fmt.Errorf("студент %q не найден", studentID)
	}
	criterion, exists := course.criteria[criterionID]
	if !exists {
		return Evaluation{}, fmt.Errorf("критерий %q отсутствует в рубрикаторе", criterionID)
	}
	score, err := NewScore(value, criterion.MaxScore())
	if err != nil {
		return Evaluation{}, err
	}
	return Evaluation{studentID, criterionID, score, strings.TrimSpace(comment)}, nil
}
