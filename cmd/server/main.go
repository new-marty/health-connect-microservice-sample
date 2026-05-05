// Package main is the health-connect HTTP server.
//
// @title           Health Connect API
// @version         1.0
// @description     Personal health data API. Aggregates Oura, Strava, Hevy, InBody, Apple Health, and meal logs. Every endpoint is also exposed as an LLM tool — see /openapi.json, /tools/openai.json, /tools/anthropic.json.
// @BasePath        /api/v1
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Bearer token. Set the same value in HEALTH_CONNECT_API_TOKEN on the server.
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-marty/health-connect/config"
	"github.com/new-marty/health-connect/internal/analysis"
	"github.com/new-marty/health-connect/internal/applehealth"
	"github.com/new-marty/health-connect/internal/database"
	"github.com/new-marty/health-connect/internal/hevy"
	"github.com/new-marty/health-connect/internal/inbody"
	"github.com/new-marty/health-connect/internal/meals"
	"github.com/new-marty/health-connect/internal/middleware"
	"github.com/new-marty/health-connect/internal/oura"
	"github.com/new-marty/health-connect/internal/spec"
	"github.com/new-marty/health-connect/internal/strava"
	"github.com/new-marty/health-connect/internal/summary"
	syncpkg "github.com/new-marty/health-connect/internal/sync"
	"github.com/new-marty/health-connect/internal/toolspec"
)

func main() {
	cfg := config.Load()

	// Initialize structured logging
	var logHandler slog.Handler
	if cfg.LogFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(logHandler))

	// Parse CORS origins
	var origins []string
	for _, o := range strings.Split(cfg.CORSOrigins, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	// Open database
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected", "path", cfg.DBPath)

	// Create repositories
	ouraRepo := oura.NewRepository(db)
	stravaRepo := strava.NewRepository(db)
	hevyRepo := hevy.NewRepository(db)
	inbodyRepo := inbody.NewRepository(db)
	appleHealthRepo := applehealth.NewRepository(db)
	mealsRepo := meals.NewRepository(db)
	syncRepo := syncpkg.NewRepository(db)

	// Create services
	ouraSvc := oura.NewService(ouraRepo)
	stravaSvc := strava.NewService(stravaRepo)
	hevySvc := hevy.NewService(hevyRepo)
	inbodySvc := inbody.NewService(inbodyRepo)
	appleHealthSvc := applehealth.NewService(appleHealthRepo)
	mealsSvc := meals.NewService(mealsRepo)

	// Create sync clients
	ouraSyncer := oura.NewSyncer(ouraRepo, cfg.OuraAccessToken)
	stravaSyncer := strava.NewSyncer(stravaRepo, db, cfg.StravaClientID, cfg.StravaClientSecret)
	hevySyncer := hevy.NewSyncer(hevyRepo, cfg.HevyAuthToken, cfg.HevyAPIKey)
	inbodySyncer := inbody.NewSyncer(inbodyRepo, cfg.InBodyLoginID, cfg.InBodyPassword)

	// Create sync scheduler
	scheduler := syncpkg.NewScheduler(syncRepo)
	scheduler.Register(ouraSyncer, cfg.OuraSyncCron)
	scheduler.Register(stravaSyncer, cfg.StravaSyncCron)
	scheduler.Register(hevySyncer, cfg.HevySyncCron)
	scheduler.Register(inbodySyncer, cfg.InBodySyncCron)
	scheduler.Start()
	defer scheduler.Stop()

	// Create summary + analysis services
	summarySvc := summary.NewService(db)
	analysisSvc := analysis.NewService(summarySvc, cfg.ClaudeAPIKey, cfg.ClaudeModel)

	// Create handlers
	ouraHandler := oura.NewHandler(ouraSvc)
	stravaHandler := strava.NewHandler(stravaSvc)
	hevyHandler := hevy.NewHandler(hevySvc)
	inbodyHandler := inbody.NewHandler(inbodySvc)
	appleHealthHandler := applehealth.NewHandler(appleHealthSvc, appleHealthRepo)
	mealsHandler := meals.NewHandler(mealsSvc)
	syncHandler := syncpkg.NewHandler(syncRepo, scheduler)
	summaryHandler := summary.NewHandler(summarySvc)
	analysisHandler := analysis.NewHandler(analysisSvc)

	// Pre-convert tool schemas at startup so /tools/* serve cached bytes.
	openaiTools, anthropicTools, err := toolspec.Convert(spec.Swagger)
	if err != nil {
		slog.Error("failed to convert openapi spec to tool schemas", "error", err)
		os.Exit(1)
	}

	// Setup router
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(origins))
	r.Use(gin.Recovery())

	// Public routes
	r.GET("/api/health", healthCheck(db))
	r.GET("/openapi.json", serveBytes("application/json", spec.Swagger))
	r.GET("/tools/openai.json", serveBytes("application/json", openaiTools))
	r.GET("/tools/anthropic.json", serveBytes("application/json", anthropicTools))

	// API v1 routes (auth-gated when HEALTH_CONNECT_API_TOKEN is set)
	api := r.Group("/api/v1")
	api.Use(middleware.BearerAuth())
	api.Use(middleware.Timeout(30 * time.Second))

	ouraHandler.RegisterRoutes(api)
	stravaHandler.RegisterRoutes(api)
	hevyHandler.RegisterRoutes(api)
	inbodyHandler.RegisterRoutes(api)
	appleHealthHandler.RegisterRoutes(api)
	mealsHandler.RegisterRoutes(api)
	syncHandler.RegisterRoutes(api)
	summaryHandler.RegisterRoutes(api)
	analysisHandler.RegisterRoutes(api)

	// Start server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		slog.Info("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}

func healthCheck(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := database.Ping(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	}
}

func serveBytes(contentType string, body []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, body)
	}
}
