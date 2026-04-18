package oura

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Syncer struct {
	repo        *Repository
	accessToken string
	client      *http.Client
}

func NewSyncer(repo *Repository, accessToken string) *Syncer {
	return &Syncer{
		repo:        repo,
		accessToken: accessToken,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Syncer) Name() string { return "oura" }

func (s *Syncer) Sync(ctx context.Context) (int, error) {
	if s.accessToken == "" {
		return 0, fmt.Errorf("oura access token not configured")
	}

	// Pull last 10 days (Oura backfills data late)
	end := time.Now().Format("2006-01-02")
	start := time.Now().AddDate(0, 0, -10).Format("2006-01-02")

	total := 0

	// Sync sleep sessions
	n, err := s.syncSleep(ctx, start, end)
	if err != nil {
		slog.Warn("oura sleep sync error", "error", err)
	}
	total += n

	// Sync daily scores (sleep + readiness + activity + spo2)
	n, err = s.syncDailyScores(ctx, start, end)
	if err != nil {
		slog.Warn("oura daily scores sync error", "error", err)
	}
	total += n

	return total, nil
}

func (s *Syncer) syncSleep(ctx context.Context, start, end string) (int, error) {
	data, err := s.fetchEndpoint(ctx, "sleep", start, end)
	if err != nil {
		return 0, err
	}

	var resp struct {
		Data []struct {
			ID                 string   `json:"id"`
			Day                string   `json:"day"`
			Type               string   `json:"type"`
			BedtimeStart       string   `json:"bedtime_start"`
			BedtimeEnd         string   `json:"bedtime_end"`
			TotalSleepDuration int      `json:"total_sleep_duration"`
			DeepSleepDuration  int      `json:"deep_sleep_duration"`
			REMSleepDuration   int      `json:"rem_sleep_duration"`
			LightSleepDuration int      `json:"light_sleep_duration"`
			AwakeTime          int      `json:"awake_time"`
			TimeInBed          int      `json:"time_in_bed"`
			Efficiency         *int     `json:"efficiency"`
			Latency            *int     `json:"latency"`
			AverageHeartRate   *float64 `json:"average_heart_rate"`
			LowestHeartRate    *int     `json:"lowest_heart_rate"`
			AverageHRV         *int     `json:"average_hrv"`
			AverageBreath      *float64 `json:"average_breath"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("parse sleep response: %w", err)
	}

	added := 0
	for _, r := range resp.Data {
		session := &SleepSession{
			OuraID:        r.ID,
			Day:           r.Day,
			Type:          r.Type,
			BedtimeStart:  r.BedtimeStart,
			BedtimeEnd:    r.BedtimeEnd,
			TotalSleepSec: r.TotalSleepDuration,
			DeepSleepSec:  r.DeepSleepDuration,
			REMSleepSec:   r.REMSleepDuration,
			LightSleepSec: r.LightSleepDuration,
			AwakeSec:      r.AwakeTime,
			TimeInBedSec:  r.TimeInBed,
			Efficiency:    r.Efficiency,
			LatencySec:    r.Latency,
			AvgHR:         r.AverageHeartRate,
			LowestHR:      r.LowestHeartRate,
			AvgHRV:        r.AverageHRV,
			AvgBreath:     r.AverageBreath,
		}
		if session.Type == "" {
			session.Type = "long_sleep"
		}
		if err := s.repo.UpsertSleep(ctx, session); err != nil {
			slog.Warn("upsert sleep error", "oura_id", r.ID, "error", err)
			continue
		}
		added++
	}
	slog.Info("oura sleep synced", "count", added)
	return added, nil
}

func (s *Syncer) syncDailyScores(ctx context.Context, start, end string) (int, error) {
	// Merge data from 4 endpoints into daily scores
	scores := map[string]*DailyScore{}

	// Daily sleep
	if data, err := s.fetchEndpoint(ctx, "daily_sleep", start, end); err == nil {
		var resp struct {
			Data []struct {
				Day          string `json:"day"`
				Score        *int   `json:"score"`
				Contributors struct {
					DeepSleep    *int `json:"deep_sleep"`
					Efficiency   *int `json:"efficiency"`
					Latency      *int `json:"latency"`
					REMSleep     *int `json:"rem_sleep"`
					Restfulness  *int `json:"restfulness"`
					Timing       *int `json:"timing"`
					TotalSleep   *int `json:"total_sleep"`
				} `json:"contributors"`
			} `json:"data"`
		}
		if json.Unmarshal(data, &resp) == nil {
			for _, r := range resp.Data {
				ds := getOrCreate(scores, r.Day)
				ds.SleepScore = r.Score
				ds.SleepDeep = r.Contributors.DeepSleep
				ds.SleepEfficiency = r.Contributors.Efficiency
				ds.SleepLatency = r.Contributors.Latency
				ds.SleepREM = r.Contributors.REMSleep
				ds.SleepRestfulness = r.Contributors.Restfulness
				ds.SleepTiming = r.Contributors.Timing
				ds.SleepTotal = r.Contributors.TotalSleep
			}
		}
	}

	// Daily readiness
	if data, err := s.fetchEndpoint(ctx, "daily_readiness", start, end); err == nil {
		var resp struct {
			Data []struct {
				Day                        string   `json:"day"`
				Score                      *int     `json:"score"`
				TemperatureDeviation       *float64 `json:"temperature_deviation"`
				TemperatureTrendDeviation  *float64 `json:"temperature_trend_deviation"`
				Contributors               struct {
					ActivityBalance    *int `json:"activity_balance"`
					BodyTemperature    *int `json:"body_temperature"`
					HRVBalance         *int `json:"hrv_balance"`
					PreviousDayActivity *int `json:"previous_day_activity"`
					PreviousNight      *int `json:"previous_night"`
					RecoveryIndex      *int `json:"recovery_index"`
					RestingHeartRate   *int `json:"resting_heart_rate"`
					SleepBalance       *int `json:"sleep_balance"`
				} `json:"contributors"`
			} `json:"data"`
		}
		if json.Unmarshal(data, &resp) == nil {
			for _, r := range resp.Data {
				ds := getOrCreate(scores, r.Day)
				ds.ReadinessScore = r.Score
				ds.ReadinessTempDev = r.TemperatureDeviation
				ds.ReadinessTempTrend = r.TemperatureTrendDeviation
				ds.ReadinessActivityBalance = r.Contributors.ActivityBalance
				ds.ReadinessBodyTemp = r.Contributors.BodyTemperature
				ds.ReadinessHRVBalance = r.Contributors.HRVBalance
				ds.ReadinessPrevDayActivity = r.Contributors.PreviousDayActivity
				ds.ReadinessPrevNight = r.Contributors.PreviousNight
				ds.ReadinessRecoveryIndex = r.Contributors.RecoveryIndex
				ds.ReadinessRestingHR = r.Contributors.RestingHeartRate
				ds.ReadinessSleepBalance = r.Contributors.SleepBalance
			}
		}
	}

	// Daily activity
	if data, err := s.fetchEndpoint(ctx, "daily_activity", start, end); err == nil {
		var resp struct {
			Data []struct {
				Day                      string `json:"day"`
				Score                    *int   `json:"score"`
				Steps                    int    `json:"steps"`
				ActiveCalories           int    `json:"active_calories"`
				TotalCalories            int    `json:"total_calories"`
				HighActivityTime         int    `json:"high_activity_time"`
				MediumActivityTime       int    `json:"medium_activity_time"`
				LowActivityTime          int    `json:"low_activity_time"`
				SedentaryTime            int    `json:"sedentary_time"`
				EquivalentWalkingDistance int    `json:"equivalent_walking_distance"`
			} `json:"data"`
		}
		if json.Unmarshal(data, &resp) == nil {
			for _, r := range resp.Data {
				ds := getOrCreate(scores, r.Day)
				ds.ActivityScore = r.Score
				ds.Steps = r.Steps
				ds.ActiveCal = r.ActiveCalories
				ds.TotalCal = r.TotalCalories
				ds.HighActivitySec = r.HighActivityTime
				ds.MediumActivitySec = r.MediumActivityTime
				ds.LowActivitySec = r.LowActivityTime
				ds.SedentarySec = r.SedentaryTime
				ds.EquivWalkingM = r.EquivalentWalkingDistance
			}
		}
	}

	// Daily SpO2
	if data, err := s.fetchEndpoint(ctx, "daily_spo2", start, end); err == nil {
		var resp struct {
			Data []struct {
				Day                     string   `json:"day"`
				SpO2Percentage          *float64 `json:"spo2_percentage"`
				BreathingDisturbanceIdx *float64 `json:"breathing_disturbance_index"`
			} `json:"data"`
		}
		if json.Unmarshal(data, &resp) == nil {
			for _, r := range resp.Data {
				ds := getOrCreate(scores, r.Day)
				ds.SpO2Pct = r.SpO2Percentage
				ds.BreathingDisturbanceIdx = r.BreathingDisturbanceIdx
			}
		}
	}

	// Upsert all daily scores
	added := 0
	for _, ds := range scores {
		if err := s.repo.UpsertScore(ctx, ds); err != nil {
			slog.Warn("upsert daily score error", "day", ds.Day, "error", err)
			continue
		}
		added++
	}
	slog.Info("oura daily scores synced", "count", added)
	return added, nil
}

func (s *Syncer) fetchEndpoint(ctx context.Context, endpoint, start, end string) ([]byte, error) {
	url := fmt.Sprintf("https://api.ouraring.com/v2/usercollection/%s?start_date=%s&end_date=%s",
		endpoint, start, end)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch %s: status %d: %s", endpoint, resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func getOrCreate(scores map[string]*DailyScore, day string) *DailyScore {
	if ds, ok := scores[day]; ok {
		return ds
	}
	ds := &DailyScore{Day: day}
	scores[day] = ds
	return ds
}
