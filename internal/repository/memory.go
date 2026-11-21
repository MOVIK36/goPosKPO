package repository

import (
	"fmt"

	"pos/internal/domain"
)

type MemoryRepository struct {
	course *domain.Course
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{course: domain.NewCourse()}
}

func (r *MemoryRepository) Load() (*domain.Course, error) {
	if r.course == nil {
		return domain.NewCourse(), nil
	}
	return r.course.Clone(), nil
}

func (r *MemoryRepository) Save(course *domain.Course) error {
	if course == nil {
		return fmt.Errorf("нельзя сохранить пустую ведомость")
	}
	r.course = course.Clone()
	return nil
}
