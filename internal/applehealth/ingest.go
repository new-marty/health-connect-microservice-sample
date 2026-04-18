package applehealth

import (
	"context"
	"log/slog"
)

// IngestPayload represents the Health Auto Export webhook payload structure.
type IngestPayload struct {
	Data struct {
		Metrics []struct {
			Name  string `json:"name"`
			Units string `json:"units"`
			Data  []struct {
				Date   string   `json:"date"`
				Qty    *float64 `json:"qty"`
				Source string   `json:"source"`
			} `json:"data"`
		} `json:"metrics"`
	} `json:"data"`
}

// Ingest processes an Apple Health Auto Export payload, writing weight to weight_readings
// and everything else to vitals.
func Ingest(ctx context.Context, repo *Repository, payload *IngestPayload) (int, error) {
	addedWeight := 0
	addedVitals := 0

	for _, m := range payload.Data.Metrics {
		for _, pt := range m.Data {
			if pt.Qty == nil {
				continue
			}
			ts := pt.Date
			date := ""
			if len(ts) >= 10 {
				date = ts[:10]
			}
			source := pt.Source
			if source == "" {
				source = "apple_health"
			}

			switch m.Name {
			case "weight_body_mass":
				w := &WeightReading{
					Timestamp: ts,
					Date:      date,
					WeightKg:  *pt.Qty,
					Source:    source,
				}
				if err := repo.InsertWeight(ctx, w); err != nil {
					slog.Debug("insert weight (likely duplicate)", "error", err)
				} else {
					addedWeight++
				}
			default:
				v := &Vital{
					Timestamp: ts,
					Date:      date,
					Metric:    m.Name,
					Value:     *pt.Qty,
					Unit:      m.Units,
					Source:    source,
				}
				if err := repo.InsertVital(ctx, v); err != nil {
					slog.Debug("insert vital (likely duplicate)", "error", err)
				} else {
					addedVitals++
				}
			}
		}
	}

	total := addedWeight + addedVitals
	if total > 0 {
		slog.Info("apple health ingested", "weight", addedWeight, "vitals", addedVitals)
	}
	return total, nil
}

