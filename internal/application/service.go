package application

import (
	"fmt"

	"pos/internal/domain"
)

type Service struct {
	repository CourseRepository
	exporter   ResultsExporter
	scorer     *domain.FinalScoreService
}

func NewService(repository CourseRepository, exporter ResultsExporter, scorer *domain.FinalScoreService) *Service {
	return &Service{repository: repository, exporter: exporter, scorer: scorer}
}

func (s *Service) change(action func(*domain.Course) error) error {
	course, err := s.repository.Load()
	if err != nil {
		return fmt.Errorf("загрузка ведомости: %w", err)
	}
	if err := action(course); err != nil {
		return err
	}
	if err := s.repository.Save(course); err != nil {
		return fmt.Errorf("сохранение ведомости: %w", err)
	}
	return nil
}

func (s *Service) AddStudent(id, name string) error {
	return s.change(func(c *domain.Course) error { return c.AddStudent(id, name) })
}

func (s *Service) RenameStudent(id, name string) error {
	return s.change(func(c *domain.Course) error { return c.RenameStudent(id, name) })
}

func (s *Service) RemoveStudent(id string) error {
	return s.change(func(c *domain.Course) error { return c.RemoveStudent(id) })
}

func (s *Service) AddCriterion(id, name string, maxScore, weight float64) error {
	return s.change(func(c *domain.Course) error { return c.AddCriterion(id, name, maxScore, weight) })
}

func (s *Service) UpdateCriterion(id, name string, maxScore, weight float64) error {
	return s.change(func(c *domain.Course) error { return c.UpdateCriterion(id, name, maxScore, weight) })
}

func (s *Service) RemoveCriterion(id string) error {
	return s.change(func(c *domain.Course) error { return c.RemoveCriterion(id) })
}

func (s *Service) SetEvaluation(studentID, criterionID string, score float64, comment string) error {
	return s.change(func(c *domain.Course) error { return c.SetEvaluation(studentID, criterionID, score, comment) })
}

func (s *Service) CancelEvaluation(studentID, criterionID string) error {
	return s.change(func(c *domain.Course) error { return c.CancelEvaluation(studentID, criterionID) })
}

func (s *Service) Students() ([]domain.Student, error) {
	course, err := s.repository.Load()
	if err != nil {
		return nil, err
	}
	return course.Students(), nil
}

func (s *Service) Criteria() ([]domain.Criterion, error) {
	course, err := s.repository.Load()
	if err != nil {
		return nil, err
	}
	return course.Criteria(), nil
}

func (s *Service) Evaluations() ([]domain.Evaluation, error) {
	course, err := s.repository.Load()
	if err != nil {
		return nil, err
	}
	return course.Evaluations(), nil
}

func (s *Service) Results() ([]domain.Result, error) {
	course, err := s.repository.Load()
	if err != nil {
		return nil, err
	}
	if len(course.Criteria()) == 0 {
		return nil, fmt.Errorf("сначала добавьте критерии оценивания")
	}
	results := make([]domain.Result, 0)
	for _, student := range course.Students() {
		result, err := s.scorer.Calculate(course, student.ID())
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) ExportResults(path string) error {
	results, err := s.Results()
	if err != nil {
		return err
	}
	return s.exporter.Export(path, results)
}
