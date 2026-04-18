package strava

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Syncer struct {
	repo         *Repository
	db           *sql.DB
	clientID     string
	clientSecret string
	client       *http.Client
}

func NewSyncer(repo *Repository, db *sql.DB, clientID, clientSecret string) *Syncer {
	return &Syncer{
		repo:         repo,
		db:           db,
		clientID:     clientID,
		clientSecret: clientSecret,
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Syncer) Name() string { return "strava" }

func (s *Syncer) Sync(ctx context.Context) (int, error) {
	if s.clientID == "" || s.clientSecret == "" {
		return 0, fmt.Errorf("strava credentials not configured")
	}

	token, err := s.getAccessToken(ctx)
	if err != nil {
		return 0, fmt.Errorf("get access token: %w", err)
	}

	// Fetch activities from last 30 days
	after := time.Now().AddDate(0, 0, -30).Unix()
	total := 0

	for page := 1; ; page++ {
		activities, err := s.fetchActivities(ctx, token, after, page)
		if err != nil {
			return total, err
		}
		if len(activities) == 0 {
			break
		}

		for _, a := range activities {
			if err := s.repo.Upsert(ctx, &a); err != nil {
				slog.Warn("upsert strava activity error", "id", a.StravaID, "error", err)
				continue
			}
			total++
		}

		if len(activities) < 100 {
			break
		}
	}

	slog.Info("strava synced", "count", total)
	return total, nil
}

func (s *Syncer) getAccessToken(ctx context.Context) (string, error) {
	// Check sync_tokens table for stored token
	var accessToken, refreshToken, expiresAt string
	err := s.db.QueryRowContext(ctx,
		"SELECT access_token, refresh_token, expires_at FROM sync_tokens WHERE provider = 'strava'",
	).Scan(&accessToken, &refreshToken, &expiresAt)

	if err == sql.ErrNoRows {
		// No stored token — try using refresh token from config to bootstrap
		return s.refreshToken(ctx, "")
	}
	if err != nil {
		return "", err
	}

	// Check if token is expired
	expiry, _ := time.Parse(time.RFC3339, expiresAt)
	if time.Now().Before(expiry) {
		return accessToken, nil
	}

	return s.refreshToken(ctx, refreshToken)
}

func (s *Syncer) refreshToken(ctx context.Context, refreshTok string) (string, error) {
	if refreshTok == "" {
		// Try to get from sync_tokens
		s.db.QueryRowContext(ctx,
			"SELECT refresh_token FROM sync_tokens WHERE provider = 'strava'",
		).Scan(&refreshTok)
	}
	if refreshTok == "" {
		return "", fmt.Errorf("no refresh token available")
	}

	data := url.Values{
		"client_id":     {s.clientID},
		"client_secret": {s.clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshTok},
	}

	resp, err := s.client.PostForm("https://www.strava.com/api/v3/oauth/token", data)
	if err != nil {
		return "", fmt.Errorf("refresh token request: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    int64  `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	// Store updated tokens
	expiresAt := time.Unix(tokenResp.ExpiresAt, 0).Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `INSERT OR REPLACE INTO sync_tokens
		(provider, access_token, refresh_token, expires_at, updated_at)
		VALUES ('strava', ?, ?, ?, datetime('now'))`,
		tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt)
	if err != nil {
		slog.Warn("failed to store strava token", "error", err)
	}

	return tokenResp.AccessToken, nil
}

func (s *Syncer) fetchActivities(ctx context.Context, token string, after int64, page int) ([]Activity, error) {
	u := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?after=%d&per_page=100&page=%d",
		after, page)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch activities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch activities: status %d: %s", resp.StatusCode, string(body))
	}

	var raw []struct {
		ID                 int64    `json:"id"`
		StartDateLocal     string   `json:"start_date_local"`
		Name               string   `json:"name"`
		Type               string   `json:"type"`
		SportType          string   `json:"sport_type"`
		Distance           float64  `json:"distance"`
		MovingTime         int      `json:"moving_time"`
		ElapsedTime        int      `json:"elapsed_time"`
		TotalElevationGain float64  `json:"total_elevation_gain"`
		AverageSpeed       *float64 `json:"average_speed"`
		MaxSpeed           *float64 `json:"max_speed"`
		HasHeartrate       bool     `json:"has_heartrate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse activities: %w", err)
	}

	activities := make([]Activity, 0, len(raw))
	for _, r := range raw {
		date := ""
		if len(r.StartDateLocal) >= 10 {
			date = r.StartDateLocal[:10]
		}
		activities = append(activities, Activity{
			StravaID:       r.ID,
			StartTime:      r.StartDateLocal,
			Date:           date,
			Name:           r.Name,
			Type:           r.Type,
			SportType:      r.SportType,
			DistanceM:      r.Distance,
			MovingTimeSec:  r.MovingTime,
			ElapsedTimeSec: r.ElapsedTime,
			ElevationGainM: r.TotalElevationGain,
			AvgSpeedMPS:    r.AverageSpeed,
			MaxSpeedMPS:    r.MaxSpeed,
			HasHeartrate:   r.HasHeartrate,
		})
	}

	return activities, nil
}
