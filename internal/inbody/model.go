package inbody

type BodyCompScan struct {
	ScanDatetime string   `json:"scan_datetime"`
	Date         string   `json:"date"`
	WeightKg     *float64 `json:"weight_kg"`
	SMMKg        *float64 `json:"smm_kg"`
	BFMKg        *float64 `json:"bfm_kg"`
	PBFPct       *float64 `json:"pbf_pct"`
	BMI          *float64 `json:"bmi"`
	BMRKcal      *int     `json:"bmr_kcal"`
	FFMKg        *float64 `json:"ffm_kg"`
	ProteinKg    *float64 `json:"protein_kg"`
	MineralKg    *float64 `json:"mineral_kg"`
	ICWKg        *float64 `json:"icw_kg"`
	ECWKg        *float64 `json:"ecw_kg"`
	VFL          *int     `json:"vfl"`
}
