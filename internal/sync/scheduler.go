package sync

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

// Syncer is the interface each source sync client must implement.
type Syncer interface {
	Name() string
	Sync(ctx context.Context) (added int, err error)
}

// Scheduler runs sync jobs on cron schedules and supports on-demand triggers.
type Scheduler struct {
	cron    *cron.Cron
	repo    *Repository
	syncers map[string]Syncer
	running atomic.Bool
}

func NewScheduler(repo *Repository) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		repo:    repo,
		syncers: make(map[string]Syncer),
	}
}

// Register adds a syncer with an optional cron schedule.
// If schedule is empty, the syncer is only triggerable via API.
func (s *Scheduler) Register(syncer Syncer, schedule string) {
	name := syncer.Name()
	s.syncers[name] = syncer

	if schedule != "" {
		_, err := s.cron.AddFunc(schedule, func() {
			s.runSync(context.Background(), name)
		})
		if err != nil {
			slog.Error("failed to register cron", "source", name, "schedule", schedule, "error", err)
		} else {
			slog.Info("registered sync schedule", "source", name, "schedule", schedule)
		}
	}
}

// Start begins the cron scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
	slog.Info("sync scheduler started")
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("sync scheduler stopped")
}

// TriggerSync runs a sync for the given source (or "all") on-demand.
func (s *Scheduler) TriggerSync(ctx context.Context, source string) (int, error) {
	if source == "all" {
		total := 0
		for name := range s.syncers {
			n := s.runSync(ctx, name)
			total += n
		}
		return total, nil
	}

	syncer, ok := s.syncers[source]
	if !ok {
		return 0, &unknownSourceError{source: source}
	}

	n, err := s.executeSync(ctx, syncer)
	return n, err
}

func (s *Scheduler) runSync(ctx context.Context, name string) int {
	syncer, ok := s.syncers[name]
	if !ok {
		return 0
	}
	n, _ := s.executeSync(ctx, syncer)
	return n
}

func (s *Scheduler) executeSync(ctx context.Context, syncer Syncer) (int, error) {
	name := syncer.Name()
	start := time.Now()
	slog.Info("sync starting", "source", name)

	added, err := syncer.Sync(ctx)

	duration := time.Since(start)
	if err != nil {
		errMsg := err.Error()
		_ = s.repo.Log(ctx, name, added, 0, "error", &errMsg)
		slog.Error("sync failed", "source", name, "duration", duration, "error", err)
		return added, err
	}

	_ = s.repo.Log(ctx, name, added, 0, "ok", nil)
	slog.Info("sync completed", "source", name, "added", added, "duration", duration)
	return added, nil
}

type unknownSourceError struct {
	source string
}

func (e *unknownSourceError) Error() string {
	return "unknown sync source: " + e.source
}
