package domain

import "fmt"

type ScorePolicy interface {
	Calculate(criteria []Criterion, scores map[string]float64) float64
	Name() string
}

type MissingAsZero struct{}

func (MissingAsZero) Name() string { return "zero" }
func (MissingAsZero) Calculate(criteria []Criterion, scores map[string]float64) float64 {
	return weightedScore(criteria, scores, true)
}

type GradedOnly struct{}

func (GradedOnly) Name() string { return "graded" }
func (GradedOnly) Calculate(criteria []Criterion, scores map[string]float64) float64 {
	return weightedScore(criteria, scores, false)
}

func weightedScore(criteria []Criterion, scores map[string]float64, includeMissing bool) float64 {
	maxWeight := 0.0
	for _, criterion := range criteria {
		_, graded := scores[criterion.ID()]
		if (graded || includeMissing) && criterion.Weight() > maxWeight {
			maxWeight = criterion.Weight()
		}
	}
	if maxWeight == 0 {
		return 0
	}
	sum, weights := 0.0, 0.0
	for _, criterion := range criteria {
		score, graded := scores[criterion.ID()]
		if !graded && !includeMissing {
			continue
		}
		weight := criterion.Weight() / maxWeight
		sum += (score / criterion.MaxScore()) * weight
		weights += weight
	}
	return 10 * (sum / weights)
}

type Result struct {
	StudentID   string
	StudentName string
	Score       float64
	Graded      int
	Total       int
	Policy      string
}

type FinalScoreService struct {
	policy ScorePolicy
}

func NewFinalScoreService(policy ScorePolicy) *FinalScoreService {
	return &FinalScoreService{policy: policy}
}

func (s *FinalScoreService) Calculate(course *Course, studentID string) (Result, error) {
	if course == nil || s.policy == nil {
		return Result{}, fmt.Errorf("не заданы ведомость или правило расчёта")
	}
	student, exists := course.Student(studentID)
	if !exists {
		return Result{}, fmt.Errorf("студент %q не найден", studentID)
	}
	criteria := course.Criteria()
	if len(criteria) == 0 {
		return Result{}, fmt.Errorf("сначала добавьте критерии оценивания")
	}
	scores := make(map[string]float64)
	for _, criterion := range criteria {
		key := evaluationKey{student.ID(), criterion.ID()}
		if evaluation, exists := course.evaluations[key]; exists {
			scores[criterion.ID()] = evaluation.Score().Value()
		}
	}
	return Result{
		StudentID: student.ID(), StudentName: student.Name(),
		Score:  s.policy.Calculate(criteria, scores),
		Graded: len(scores), Total: len(criteria), Policy: s.policy.Name(),
	}, nil
}
