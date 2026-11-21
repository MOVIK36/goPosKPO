package di_test

import (
	"testing"

	"pos/internal/di"
	"pos/internal/domain"
	"pos/internal/export"
	"pos/internal/repository"
)

type fixedPolicy struct{}

func (fixedPolicy) Calculate([]domain.Criterion, map[string]float64) float64 { return 4.2 }
func (fixedPolicy) Name() string                                             { return "test" }

func TestContainerInjectsPolicy(t *testing.T) {
	container, err := di.NewContainer(repository.NewMemoryRepository(), export.CSVExporter{}, fixedPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if err := container.Service.AddStudent("s1", "Анна"); err != nil {
		t.Fatal(err)
	}
	if err := container.Service.AddCriterion("hw", "Работа", 10, 1); err != nil {
		t.Fatal(err)
	}
	results, err := container.Service.Results()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Score != 4.2 || results[0].Policy != "test" {
		t.Fatal("подставленная стратегия не использовалась")
	}
}

func TestContainerRejectsMissingDependencies(t *testing.T) {
	if _, err := di.NewDefaultContainer("unknown"); err == nil {
		t.Fatal("принято неизвестное правило")
	}
	if _, err := di.NewContainer(nil, export.CSVExporter{}, domain.MissingAsZero{}); err == nil {
		t.Fatal("принят nil-репозиторий")
	}
	if _, err := di.NewContainer(repository.NewMemoryRepository(), nil, domain.MissingAsZero{}); err == nil {
		t.Fatal("принят nil-экспортёр")
	}
	if _, err := di.NewContainer(repository.NewMemoryRepository(), export.CSVExporter{}, nil); err == nil {
		t.Fatal("принята nil-стратегия")
	}
}
