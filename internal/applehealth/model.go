package applehealth

type WeightReading struct {
	Timestamp  string   `json:"timestamp"`
	Date       string   `json:"date"`
	WeightKg   float64  `json:"weight_kg"`
	BodyFatPct *float64 `json:"body_fat_pct"`
	BMI        *float64 `json:"bmi"`
	LeanMassKg *float64 `json:"lean_mass_kg"`
	Source     string   `json:"source"`
}

type Vital struct {
	ID        int64   `json:"id"`
	Timestamp string  `json:"timestamp"`
	Date      string  `json:"date"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Source    string  `json:"source"`
}
