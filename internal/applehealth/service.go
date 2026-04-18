package applehealth

import (
	"context"
	"time"

	"github.com/new-marty/health-connect/internal/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListWeight(ctx context.Context, from, to string) ([]WeightReading, error) {
	return s.repo.ListWeight(ctx, from, to)
}

func (s *Service) CreateWeight(ctx context.Context, w *WeightReading) error {
	if w.WeightKg <= 0 {
		return apperror.InvalidInput("weight_kg must be positive")
	}
	if w.Date == "" {
		w.Date = time.Now().Format("2006-01-02")
	}
	if w.Timestamp == "" {
		w.Timestamp = time.Now().Format(time.RFC3339)
	}
	if w.Source == "" {
		w.Source = "manual"
	}
	return s.repo.InsertWeight(ctx, w)
}

func (s *Service) ListVitals(ctx context.Context, metric, from, to string) ([]Vital, error) {
	return s.repo.ListVitals(ctx, metric, from, to)
}
