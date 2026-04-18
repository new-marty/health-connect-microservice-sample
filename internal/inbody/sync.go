package inbody

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const apiBase = "https://appapijpnv3.lookinbody.com/V2"

type Syncer struct {
	repo     *Repository
	loginID  string
	password string
	client   *http.Client
}

func NewSyncer(repo *Repository, loginID, password string) *Syncer {
	return &Syncer{
		repo:     repo,
		loginID:  loginID,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Syncer) Name() string { return "inbody" }

func (s *Syncer) Sync(ctx context.Context) (int, error) {
	if s.loginID == "" || s.password == "" {
		return 0, fmt.Errorf("inbody credentials not configured")
	}

	token, uid, err := s.login(ctx)
	if err != nil {
		return 0, fmt.Errorf("inbody login: %w", err)
	}

	total := 0
	pageSize := 20
	for index := 0; ; index += pageSize {
		scans, err := s.fetchPage(ctx, token, uid, index, pageSize)
		if err != nil {
			return total, err
		}
		if len(scans) == 0 {
			break
		}

		for _, scan := range scans {
			if err := s.repo.Upsert(ctx, &scan); err != nil {
				slog.Warn("upsert inbody scan error", "datetime", scan.ScanDatetime, "error", err)
				continue
			}
			total++
		}

		if len(scans) < pageSize {
			break
		}
	}

	slog.Info("inbody synced", "count", total)
	return total, nil
}

func (s *Syncer) login(ctx context.Context) (string, string, error) {
	payload := map[string]string{
		"LoginID":                    s.loginID,
		"LoginPW":                    s.password,
		"SyncDatetime":              "1990-01-01 11:11:11",
		"SyncDatetimeInBody":        "1990-01-01 11:11:11",
		"SyncDatetimeExercise":      "2990-01-01 11:11:11",
		"SyncDatetimeNutrition":     "2990-01-01 11:11:11",
		"SyncDatetimeSleep":         "2990-01-01 11:11:11",
		"SyncDatetimeBasalMedical":  "2990-01-01 11:11:11",
		"SyncDatetimeCardiac":       "2990-01-01 11:11:11",
		"CountryCode":               "81",
		"SyncType":                  "Main;InBody",
		"Type":                      "AutoLogin",
		"DeviceType":                "iPhone14,5 iPhone 13 26.0 (390*844)",
		"LogCode":                   s.loginID,
		"AppType":                   "IOS",
		"AppVersion":                "2.9.9",
		"OSVersion":                 "26.0",
		"PhoneModel":                "iPhone14,5 iPhone 13",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST",
		apiBase+"/main/GetLoginWithSyncDataPartV2", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "InBody/2.9.9 (iPhone; iOS 26.0; Scale/3.00)")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"Token"`
		Data  struct {
			UID string `json:"UID"`
		} `json:"Data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("parse login response: %w", err)
	}
	if result.Token == "" {
		return "", "", fmt.Errorf("login failed: empty token")
	}

	return result.Token, result.Data.UID, nil
}

func (s *Syncer) fetchPage(ctx context.Context, token, uid string, index, pageSize int) ([]BodyCompScan, error) {
	payload := map[string]interface{}{
		"uid":             uid,
		"syncDatetime":    "1990-01-01 11:11:11",
		"NumberPerData":   fmt.Sprintf("%d", pageSize),
		"CurrentIndex":    fmt.Sprintf("%d", index),
		"Language":        "en-US",
		"LogCode":         s.loginID,
		"AppType":         "IOS",
		"AppVersion":      "2.9.9",
		"OSVersion":       "26.0",
		"PhoneModel":      "iPhone14,5 iPhone 13",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST",
		apiBase+"/InBody/GetInBodyData", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "InBody/2.9.9 (iPhone; iOS 26.0; Scale/3.00)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch inbody page: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data json.RawMessage `json:"Data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse inbody response: %w", err)
	}

	var rawScans []struct {
		DATETIMES string `json:"DATETIMES"`
		BCA       struct {
			WT      interface{} `json:"WT"`
			WEIGHT  interface{} `json:"WEIGHT"`
			SLM     interface{} `json:"SLM"`
			SMM     interface{} `json:"SMM"`
			BFM     interface{} `json:"BFM"`
			PBF     interface{} `json:"PBF"`
			PBFM    interface{} `json:"PBFM"`
			BMI     interface{} `json:"BMI"`
			BMR     interface{} `json:"BMR"`
			FFM     interface{} `json:"FFM"`
			PROTEIN interface{} `json:"PROTEIN"`
			MINERAL interface{} `json:"MINERAL"`
			ICW     interface{} `json:"ICW"`
			ECW     interface{} `json:"ECW"`
			VFL     interface{} `json:"VFL"`
		} `json:"BCA"`
	}
	if err := json.Unmarshal(result.Data, &rawScans); err != nil {
		return nil, nil // Data might not be an array
	}

	scans := make([]BodyCompScan, 0, len(rawScans))
	for _, r := range rawScans {
		dt := r.DATETIMES
		if len(dt) < 8 {
			continue
		}
		date := fmt.Sprintf("%s-%s-%s", dt[:4], dt[4:6], dt[6:8])

		scan := BodyCompScan{
			ScanDatetime: dt,
			Date:         date,
			WeightKg:     toFloat(coalesce(r.BCA.WT, r.BCA.WEIGHT)),
			SMMKg:        toFloat(coalesce(r.BCA.SLM, r.BCA.SMM)),
			BFMKg:        toFloat(r.BCA.BFM),
			PBFPct:       toFloat(coalesce(r.BCA.PBF, r.BCA.PBFM)),
			BMI:          toFloat(r.BCA.BMI),
			BMRKcal:      toInt(r.BCA.BMR),
			FFMKg:        toFloat(r.BCA.FFM),
			ProteinKg:    toFloat(r.BCA.PROTEIN),
			MineralKg:    toFloat(r.BCA.MINERAL),
			ICWKg:        toFloat(r.BCA.ICW),
			ECWKg:        toFloat(r.BCA.ECW),
			VFL:          toInt(r.BCA.VFL),
		}
		scans = append(scans, scan)
	}

	return scans, nil
}

func toFloat(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		if val == 0 {
			return nil
		}
		return &val
	case string:
		return nil
	default:
		return nil
	}
}

func toInt(v interface{}) *int {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		if val == 0 {
			return nil
		}
		i := int(val)
		return &i
	default:
		return nil
	}
}

func coalesce(a, b interface{}) interface{} {
	if a != nil {
		return a
	}
	return b
}
