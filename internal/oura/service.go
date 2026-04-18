package oura

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListSleep(ctx context.Context, from, to, typ string) ([]SleepSession, error) {
	return s.repo.ListSleep(ctx, from, to, typ)
}

func (s *Service) GetSleepByDay(ctx context.Context, day string) ([]SleepSession, error) {
	return s.repo.GetSleepByDay(ctx, day)
}

func (s *Service) ListScores(ctx context.Context, from, to string) ([]DailyScore, error) {
	return s.repo.ListScores(ctx, from, to)
}

func (s *Service) GetScoreByDay(ctx context.Context, day string) (*DailyScore, error) {
	return s.repo.GetScoreByDay(ctx, day)
}
