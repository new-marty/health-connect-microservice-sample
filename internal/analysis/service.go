package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/new-marty/health-connect/internal/summary"
)

type Service struct {
	summarySvc *summary.Service
	apiKey     string
	model      string
	client     *http.Client
}

func NewService(summarySvc *summary.Service, apiKey, model string) *Service {
	return &Service{
		summarySvc: summarySvc,
		apiKey:     apiKey,
		model:      model,
		client:     &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *Service) Analyze(ctx context.Context, date string) (*AnalysisResponse, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("claude API key not configured")
	}

	if date == "" {
		date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	// Get summary data for the date
	summaries, err := s.summarySvc.DailySummary(ctx, "", "", date)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}

	prompt := buildPrompt(summaries, date)

	analysis, err := s.callClaude(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("claude API: %w", err)
	}

	return &AnalysisResponse{
		Date:     date,
		Analysis: analysis,
	}, nil
}

func buildPrompt(summaries []summary.DailySummary, date string) string {
	var parts []string
	parts = append(parts,
		"You are a concise health analyst. Analyze this data and give 3-4 bullet points of SPECIFIC insights.",
		"Use ONLY the data provided. No generic tips. Connect domains where data supports it.",
		"Format: Plain text only. No markdown. Max 180 words.",
		"",
	)

	if len(summaries) > 0 {
		ds := summaries[0]
		parts = append(parts, fmt.Sprintf("DATE: %s", ds.Date))

		if ds.SleepScore != nil {
			line := fmt.Sprintf("  Sleep score: %d", *ds.SleepScore)
			if ds.SleepHrs != nil {
				line += fmt.Sprintf(", duration: %.1fh", *ds.SleepHrs)
			}
			parts = append(parts, line)
		}
		if ds.ReadinessScore != nil {
			parts = append(parts, fmt.Sprintf("  Readiness: %d", *ds.ReadinessScore))
		}
		if ds.SleepHRV != nil {
			line := fmt.Sprintf("  HRV: %dms", *ds.SleepHRV)
			if ds.SleepHR != nil {
				line += fmt.Sprintf(", RHR: %.0fbpm", *ds.SleepHR)
			}
			parts = append(parts, line)
		}
		if ds.Steps != nil {
			parts = append(parts, fmt.Sprintf("  Steps: %d", *ds.Steps))
		}
		if ds.ActiveCal != nil {
			parts = append(parts, fmt.Sprintf("  Active calories: %d", *ds.ActiveCal))
		}
		if ds.CaloriesIn != nil {
			parts = append(parts, fmt.Sprintf("  Calories in: %d", *ds.CaloriesIn))
		}
		if ds.ProteinG != nil {
			parts = append(parts, fmt.Sprintf("  Protein: %.0fg", *ds.ProteinG))
		}
		if ds.WeightKg != nil {
			parts = append(parts, fmt.Sprintf("  Weight: %.1fkg", *ds.WeightKg))
		}
		if ds.WorkoutCount > 0 {
			parts = append(parts, fmt.Sprintf("  Workouts: %d", ds.WorkoutCount))
		}
	} else {
		parts = append(parts, "No data available for this date.")
	}

	return strings.Join(parts, "\n")
}

func (s *Service) callClaude(ctx context.Context, prompt string) (string, error) {
	body := map[string]interface{}{
		"model":      s.model,
		"max_tokens": 300,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API error: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return result.Content[0].Text, nil
}
