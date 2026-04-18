package hevy

type Workout struct {
	ID          int64        `json:"id"`
	Date        string       `json:"date"`
	SessionType string       `json:"session_type"`
	SessionNum  int          `json:"session_num"`
	BodyweightKg *float64    `json:"bodyweight_kg"`
	Notes       *string      `json:"notes"`
	Source      string       `json:"source"`
	Sets        []WorkoutSet `json:"sets,omitempty"`
}

type WorkoutSet struct {
	ID           int64    `json:"id"`
	WorkoutID    int64    `json:"workout_id"`
	ExerciseName string   `json:"exercise_name"`
	SetNum       int      `json:"set_num"`
	WeightKg     float64  `json:"weight_kg"`
	Reps         int      `json:"reps"`
	RPE          *float64 `json:"rpe"`
	Notes        *string  `json:"notes"`
}
