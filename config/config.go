package config

import "os"

type Config struct {
	Port        string
	DBPath      string
	LogFormat   string
	CORSOrigins string

	// External API credentials
	OuraAccessToken string
	InBodyLoginID   string
	InBodyPassword  string
	HevyAuthToken   string
	HevyAPIKey      string

	// Strava OAuth
	StravaClientID     string
	StravaClientSecret string
	StravaRefreshToken string

	// AI
	ClaudeAPIKey string
	ClaudeModel  string

	// Sync schedules (cron expressions)
	OuraSyncCron   string
	HevySyncCron   string
	StravaSyncCron string
	InBodySyncCron string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBPath:      getEnv("DB_PATH", "./data/health.db"),
		LogFormat:   getEnv("LOG_FORMAT", "text"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),

		OuraAccessToken: getEnv("OURA_ACCESS_TOKEN", ""),
		InBodyLoginID:   getEnv("INBODY_LOGIN_ID", ""),
		InBodyPassword:  getEnv("INBODY_PASSWORD", ""),
		HevyAuthToken:   getEnv("HEVY_AUTH_TOKEN", ""),
		HevyAPIKey:      getEnv("HEVY_API_KEY", ""),

		StravaClientID:     getEnv("STRAVA_CLIENT_ID", ""),
		StravaClientSecret: getEnv("STRAVA_CLIENT_SECRET", ""),
		StravaRefreshToken: getEnv("STRAVA_REFRESH_TOKEN", ""),

		ClaudeAPIKey: getEnv("CLAUDE_API_KEY", ""),
		ClaudeModel:  getEnv("CLAUDE_MODEL", "claude-sonnet-4-20250514"),

		OuraSyncCron:   getEnv("OURA_SYNC_CRON", "0 */6 * * *"),
		HevySyncCron:   getEnv("HEVY_SYNC_CRON", "0 */4 * * *"),
		StravaSyncCron: getEnv("STRAVA_SYNC_CRON", "0 */6 * * *"),
		InBodySyncCron: getEnv("INBODY_SYNC_CRON", "0 8 * * 0"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
