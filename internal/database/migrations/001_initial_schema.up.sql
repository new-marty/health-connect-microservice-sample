-- Sync log
CREATE TABLE IF NOT EXISTS sync_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now')),
    records_added INTEGER NOT NULL DEFAULT 0,
    records_updated INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'ok',
    error_msg TEXT
);

-- Oura: sleep sessions
CREATE TABLE IF NOT EXISTS sleep_sessions (
    oura_id TEXT PRIMARY KEY,
    day TEXT NOT NULL,
    type TEXT NOT NULL,
    bedtime_start TEXT NOT NULL,
    bedtime_end TEXT NOT NULL,
    total_sleep_sec INTEGER NOT NULL DEFAULT 0,
    deep_sleep_sec INTEGER NOT NULL DEFAULT 0,
    rem_sleep_sec INTEGER NOT NULL DEFAULT 0,
    light_sleep_sec INTEGER NOT NULL DEFAULT 0,
    awake_sec INTEGER NOT NULL DEFAULT 0,
    time_in_bed_sec INTEGER NOT NULL DEFAULT 0,
    efficiency INTEGER,
    latency_sec INTEGER,
    avg_hr REAL,
    lowest_hr INTEGER,
    avg_hrv INTEGER,
    avg_breath REAL
);

CREATE INDEX IF NOT EXISTS idx_sleep_day_type ON sleep_sessions(day, type);
CREATE INDEX IF NOT EXISTS idx_sleep_bedtime_start ON sleep_sessions(bedtime_start);

-- Oura: daily scores (sleep + readiness + activity + SpO2)
CREATE TABLE IF NOT EXISTS daily_scores (
    day TEXT PRIMARY KEY,
    sleep_score INTEGER,
    sleep_deep INTEGER,
    sleep_efficiency INTEGER,
    sleep_latency INTEGER,
    sleep_rem INTEGER,
    sleep_restfulness INTEGER,
    sleep_timing INTEGER,
    sleep_total INTEGER,
    readiness_score INTEGER,
    readiness_temp_dev REAL,
    readiness_temp_trend REAL,
    readiness_activity_balance INTEGER,
    readiness_body_temp INTEGER,
    readiness_hrv_balance INTEGER,
    readiness_prev_day_activity INTEGER,
    readiness_prev_night INTEGER,
    readiness_recovery_index INTEGER,
    readiness_resting_hr INTEGER,
    readiness_sleep_balance INTEGER,
    activity_score INTEGER,
    steps INTEGER NOT NULL DEFAULT 0,
    active_cal INTEGER NOT NULL DEFAULT 0,
    total_cal INTEGER NOT NULL DEFAULT 0,
    high_activity_sec INTEGER NOT NULL DEFAULT 0,
    medium_activity_sec INTEGER NOT NULL DEFAULT 0,
    low_activity_sec INTEGER NOT NULL DEFAULT 0,
    sedentary_sec INTEGER NOT NULL DEFAULT 0,
    equivalent_walking_m INTEGER NOT NULL DEFAULT 0,
    spo2_pct REAL,
    breathing_disturbance_idx REAL
);

CREATE INDEX IF NOT EXISTS idx_daily_scores_day ON daily_scores(day);

-- Strava: activities
CREATE TABLE IF NOT EXISTS activities (
    strava_id INTEGER PRIMARY KEY,
    start_time TEXT NOT NULL,
    date TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    sport_type TEXT NOT NULL,
    distance_m REAL NOT NULL DEFAULT 0,
    moving_time_sec INTEGER NOT NULL DEFAULT 0,
    elapsed_time_sec INTEGER NOT NULL DEFAULT 0,
    elevation_gain_m REAL NOT NULL DEFAULT 0,
    avg_speed_mps REAL,
    max_speed_mps REAL,
    has_heartrate INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_activities_date ON activities(date);
CREATE INDEX IF NOT EXISTS idx_activities_type_date ON activities(type, date);

-- Hevy: workouts + sets
CREATE TABLE IF NOT EXISTS workouts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    session_type TEXT NOT NULL,
    session_num INTEGER NOT NULL DEFAULT 1,
    bodyweight_kg REAL,
    notes TEXT,
    source TEXT NOT NULL DEFAULT 'manual',
    UNIQUE(date, session_type, session_num)
);

CREATE TABLE IF NOT EXISTS workout_sets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_name TEXT NOT NULL,
    set_num INTEGER NOT NULL,
    weight_kg REAL NOT NULL DEFAULT 0,
    reps INTEGER NOT NULL DEFAULT 0,
    rpe REAL,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_workouts_date ON workouts(date);
CREATE INDEX IF NOT EXISTS idx_sets_workout ON workout_sets(workout_id);
CREATE INDEX IF NOT EXISTS idx_sets_exercise ON workout_sets(exercise_name);

-- InBody: body composition scans
CREATE TABLE IF NOT EXISTS body_comp_scans (
    scan_datetime TEXT PRIMARY KEY,
    date TEXT NOT NULL,
    weight_kg REAL,
    smm_kg REAL,
    bfm_kg REAL,
    pbf_pct REAL,
    bmi REAL,
    bmr_kcal INTEGER,
    ffm_kg REAL,
    protein_kg REAL,
    mineral_kg REAL,
    icw_kg REAL,
    ecw_kg REAL,
    vfl INTEGER
);

CREATE INDEX IF NOT EXISTS idx_body_comp_date ON body_comp_scans(date);

-- Apple Health: weight readings
CREATE TABLE IF NOT EXISTS weight_readings (
    timestamp TEXT NOT NULL,
    date TEXT NOT NULL,
    weight_kg REAL NOT NULL,
    body_fat_pct REAL,
    bmi REAL,
    lean_mass_kg REAL,
    source TEXT NOT NULL,
    UNIQUE(timestamp, source)
);

CREATE INDEX IF NOT EXISTS idx_weight_date ON weight_readings(date);
CREATE INDEX IF NOT EXISTS idx_weight_source_date ON weight_readings(source, date);

-- Apple Health: vitals (EAV)
CREATE TABLE IF NOT EXISTS vitals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    date TEXT NOT NULL,
    metric TEXT NOT NULL,
    value REAL NOT NULL,
    unit TEXT NOT NULL,
    source TEXT NOT NULL,
    UNIQUE(timestamp, metric, source)
);

CREATE INDEX IF NOT EXISTS idx_vitals_metric_date ON vitals(metric, date);
CREATE INDEX IF NOT EXISTS idx_vitals_date_metric ON vitals(date, metric);

-- Meals
CREATE TABLE IF NOT EXISTS meals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    meal TEXT NOT NULL,
    description TEXT NOT NULL,
    calories INTEGER NOT NULL DEFAULT 0,
    protein_g REAL NOT NULL DEFAULT 0,
    fat_g REAL NOT NULL DEFAULT 0,
    carbs_g REAL NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'manual',
    UNIQUE(date, meal, description)
);

CREATE INDEX IF NOT EXISTS idx_meals_date ON meals(date);

-- Sync tokens (for Strava OAuth etc.)
CREATE TABLE IF NOT EXISTS sync_tokens (
    provider TEXT PRIMARY KEY,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TEXT,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
