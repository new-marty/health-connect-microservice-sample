package inbody

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, from, to string, limit int) ([]BodyCompScan, error) {
	return s.repo.List(ctx, from, to, limit)
}

func (s *Service) Latest(ctx context.Context) (*BodyCompScan, error) {
	return s.repo.Latest(ctx)
}
