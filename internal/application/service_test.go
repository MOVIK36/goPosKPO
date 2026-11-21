package application_test

import (
	"errors"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/repository"
)

type failingRepository struct {
	loadErr, saveErr error
}

func (r failingRepository) Load() (*domain.Course, error) {
	return domain.NewCourse(), r.loadErr
}

func (r failingRepository) Save(*domain.Course) error { return r.saveErr }

type fakeExporter struct {
	path    string
	results []domain.Result
	err     error
}

func (e *fakeExporter) Export(path string, results []domain.Result) error {
	e.path, e.results = path, results
	return e.err
}

func TestRepositoryErrorsAreReturned(t *testing.T) {
	failure := errors.New("хранилище недоступно")
	scorer := domain.NewFinalScoreService(domain.MissingAsZero{})
	for _, repo := range []application.CourseRepository{
		failingRepository{loadErr: failure}, failingRepository{saveErr: failure},
	} {
		service := application.NewService(repo, &fakeExporter{}, scorer)
		if err := service.AddStudent("s1", "Анна"); !errors.Is(err, failure) {
			t.Fatalf("ошибка хранилища потеряна: %v", err)
		}
	}
	service := application.NewService(failingRepository{loadErr: failure}, &fakeExporter{}, scorer)
	if _, err := service.Students(); !errors.Is(err, failure) {
		t.Fatal("потеряна ошибка чтения студентов")
	}
	if _, err := service.Criteria(); !errors.Is(err, failure) {
		t.Fatal("потеряна ошибка чтения критериев")
	}
	if _, err := service.Evaluations(); !errors.Is(err, failure) {
		t.Fatal("потеряна ошибка чтения оценок")
	}
	if err := service.ExportResults("ignored.csv"); !errors.Is(err, failure) {
		t.Fatal("экспорт проигнорировал ошибку чтения")
	}
}

func TestExportUsesCurrentResultsAndPropagatesError(t *testing.T) {
	exporter := &fakeExporter{}
	service := application.NewService(repository.NewMemoryRepository(), exporter, domain.NewFinalScoreService(domain.MissingAsZero{}))
	if err := service.ExportResults("ignored.csv"); err == nil {
		t.Fatal("разрешён экспорт без рубрикатора")
	}
	for _, err := range []error{
		service.AddStudent("s1", "Анна"), service.AddCriterion("hw", "Работа", 10, 1),
		service.SetEvaluation("s1", "hw", 8, ""),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := service.ExportResults("chosen.csv"); err != nil {
		t.Fatal(err)
	}
	if exporter.path != "chosen.csv" || len(exporter.results) != 1 || exporter.results[0].Score != 8 {
		t.Fatalf("неверные данные экспортёра: %+v", exporter)
	}
	failure := errors.New("ошибка записи")
	exporter.err = failure
	if err := service.ExportResults("chosen.csv"); !errors.Is(err, failure) {
		t.Fatal("ошибка экспортёра потеряна")
	}
}
