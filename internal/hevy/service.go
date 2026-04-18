package hevy

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

func (s *Service) List(ctx context.Context, from, to string) ([]Workout, error) {
	return s.repo.List(ctx, from, to)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Workout, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, apperror.NotFound("workout")
	}
	return w, nil
}
