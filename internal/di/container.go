package di

import (
	"fmt"

	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/export"
	"pos/internal/repository"
)

type Container struct {
	Service *application.Service
}

func NewContainer(repo application.CourseRepository, exporter application.ResultsExporter, policy domain.ScorePolicy) (*Container, error) {
	if repo == nil || exporter == nil || policy == nil {
		return nil, fmt.Errorf("для контейнера нужны репозиторий, экспортёр и стратегия")
	}
	scorer := domain.NewFinalScoreService(policy)
	return &Container{Service: application.NewService(repo, exporter, scorer)}, nil
}

func NewDefaultContainer(policyName string) (*Container, error) {
	var policy domain.ScorePolicy
	switch policyName {
	case "zero":
		policy = domain.MissingAsZero{}
	case "graded":
		policy = domain.GradedOnly{}
	default:
		return nil, fmt.Errorf("неизвестная стратегия %q: выберите zero или graded", policyName)
	}
	return NewContainer(repository.NewMemoryRepository(), export.CSVExporter{}, policy)
}
