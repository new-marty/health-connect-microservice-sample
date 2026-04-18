package meals

type Meal struct {
	ID          int64   `json:"id"`
	Date        string  `json:"date"`
	Meal        string  `json:"meal"`
	Description string  `json:"description"`
	Calories    int     `json:"calories"`
	ProteinG    float64 `json:"protein_g"`
	FatG        float64 `json:"fat_g"`
	CarbsG      float64 `json:"carbs_g"`
	Source      string  `json:"source"`
}
