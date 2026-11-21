package application

import "pos/internal/domain"

type CourseRepository interface {
	Load() (*domain.Course, error)
	Save(course *domain.Course) error
}

type ResultsExporter interface {
	Export(path string, results []domain.Result) error
}
