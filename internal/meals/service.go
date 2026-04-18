package meals

import (
	"context"

	"github.com/new-marty/health-connect/internal/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, date, from, to string) ([]Meal, error) {
	return s.repo.List(ctx, date, from, to)
}

func (s *Service) Create(ctx context.Context, m *Meal) (int64, error) {
	if m.Date == "" {
		return 0, apperror.InvalidInput("date is required")
	}
	if m.Description == "" {
		return 0, apperror.InvalidInput("description is required")
	}
	if m.Meal == "" {
		m.Meal = "unknown"
	}
	if m.Source == "" {
		m.Source = "manual"
	}
	return s.repo.Create(ctx, m)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return apperror.NotFound("meal")
	}
	return nil
}
