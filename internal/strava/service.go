package strava

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

func (s *Service) List(ctx context.Context, from, to, typ string) ([]Activity, error) {
	return s.repo.List(ctx, from, to, typ)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Activity, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, apperror.NotFound("activity")
	}
	return a, nil
}
