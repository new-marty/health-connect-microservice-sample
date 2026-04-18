package hevy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Syncer struct {
	repo      *Repository
	authToken string
	apiKey    string
	client    *http.Client
}

func NewSyncer(repo *Repository, authToken, apiKey string) *Syncer {
	return &Syncer{
		repo:      repo,
		authToken: authToken,
		apiKey:    apiKey,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Syncer) Name() string { return "hevy" }

func (s *Syncer) Sync(ctx context.Context) (int, error) {
	if s.authToken == "" || s.apiKey == "" {
		return 0, fmt.Errorf("hevy credentials not configured")
	}

	total := 0
	var cursor int64

	for page := 0; page < 20; page++ {
		workouts, nextCursor, hasMore, err := s.fetchPage(ctx, cursor)
		if err != nil {
			return total, err
		}

		for _, w := range workouts {
			n, err := s.upsertWorkout(ctx, &w)
			if err != nil {
				slog.Warn("upsert hevy workout error", "id", w.ID, "error", err)
				continue
			}
			total += n
		}

		if !hasMore || nextCursor == cursor {
			break
		}
		cursor = nextCursor
	}

	slog.Info("hevy synced", "count", total)
	return total, nil
}

type hevyAPIWorkout struct {
	ID        string `json:"id"`
	Index     int64  `json:"index"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	StartTime int64  `json:"start_time"`
	Exercises []struct {
		Title string `json:"title"`
		Sets  []struct {
			Type     string   `json:"type"`
			WeightKg float64  `json:"weight_kg"`
			Reps     int      `json:"reps"`
			RPE      *float64 `json:"rpe"`
		} `json:"sets"`
	} `json:"exercises"`
}

func (s *Syncer) fetchPage(ctx context.Context, since int64) ([]hevyAPIWorkout, int64, bool, error) {
	body, err := json.Marshal(map[string]int64{"since": since})
	if err != nil {
		return nil, 0, false, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.hevyapp.com/workouts_sync_batch", bytes.NewReader(body))
	if err != nil {
		return nil, 0, false, err
	}
	req.Header.Set("auth-token", s.authToken)
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, false, fmt.Errorf("fetch hevy page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, 0, false, fmt.Errorf("hevy API error: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Updated []hevyAPIWorkout `json:"updated"`
		IsMore  bool             `json:"isMore"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, false, fmt.Errorf("parse hevy response: %w", err)
	}

	var nextCursor int64
	if len(result.Updated) > 0 {
		nextCursor = result.Updated[len(result.Updated)-1].Index
	}

	return result.Updated, nextCursor, result.IsMore, nil
}

func (s *Syncer) upsertWorkout(ctx context.Context, hw *hevyAPIWorkout) (int, error) {
	jst := time.FixedZone("JST", 9*3600)
	dt := time.Unix(hw.StartTime, 0).In(jst)
	date := dt.Format("2006-01-02")

	title := hw.Name
	if title == "" {
		title = hw.Title
	}
	sessionType := classifySession(title)

	w := &Workout{
		Date:        date,
		SessionType: sessionType,
		SessionNum:  1,
		Notes:       &title,
		Source:      "hevy",
	}

	workoutID, err := s.repo.UpsertWorkout(ctx, w)
	if err != nil {
		return 0, err
	}

	// Delete existing sets and re-insert
	if err := s.repo.DeleteSetsByWorkout(ctx, workoutID); err != nil {
		return 0, err
	}

	setCount := 0
	for _, ex := range hw.Exercises {
		setNum := 0
		for _, set := range ex.Sets {
			if set.Type == "warmup" {
				continue
			}
			setNum++
			ws := &WorkoutSet{
				WorkoutID:    workoutID,
				ExerciseName: ex.Title,
				SetNum:       setNum,
				WeightKg:     set.WeightKg,
				Reps:         set.Reps,
				RPE:          set.RPE,
			}
			if err := s.repo.InsertSet(ctx, ws); err != nil {
				slog.Warn("insert set error", "exercise", ex.Title, "error", err)
				continue
			}
			setCount++
		}
	}

	return 1, nil
}

func classifySession(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "chest"):
		return "chest"
	case strings.Contains(lower, "back"):
		return "back"
	case strings.Contains(lower, "leg"):
		return "legs"
	case strings.Contains(lower, "shoulder"):
		return "shoulders"
	case strings.Contains(lower, "arm"):
		return "arms"
	default:
		return "general"
	}
}
