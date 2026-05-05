package applehealth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-marty/health-connect/internal/apperror"
)

type Handler struct {
	svc  *Service
	repo *Repository
}

func NewHandler(svc *Service, repo *Repository) *Handler {
	return &Handler{svc: svc, repo: repo}
}

// ListWeight godoc
// @Summary      List weight readings
// @Description  Returns weight readings (kg) and optional body composition fields over an optional date range. Sourced from Apple Health, manual entries, or InBody pushes.
// @Tags         apple-health
// @Produce      json
// @Security     BearerAuth
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Success      200  {object}  object{data=[]applehealth.WeightReading}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /apple-health/weight [get]
// @ID           applehealth_list_weight
func (h *Handler) ListWeight(c *gin.Context) {
	readings, err := h.svc.ListWeight(c.Request.Context(), c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": readings})
}

// CreateWeightRequest is the JSON body for logging a weight reading.
type CreateWeightRequest struct {
	WeightKg   float64  `json:"weight_kg" binding:"required" example:"72.5"`
	BodyFatPct *float64 `json:"body_fat_pct,omitempty" example:"18.4"`
	BMI        *float64 `json:"bmi,omitempty" example:"22.1"`
	LeanMassKg *float64 `json:"lean_mass_kg,omitempty" example:"58.2"`
	Date       string   `json:"date,omitempty" example:"2026-05-04"`
	Source     string   `json:"source,omitempty" example:"manual"`
}

// CreateWeight godoc
// @Summary      Log a weight reading
// @Description  Records a manual or device-sourced weight reading. Use this when the user wants to log their weight or body composition.
// @Tags         apple-health
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  applehealth.CreateWeightRequest  true  "Weight reading"
// @Success      201  {object}  object{status=string}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /apple-health/weight [post]
// @ID           applehealth_create_weight
func (h *Handler) CreateWeight(c *gin.Context) {
	var req CreateWeightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondGin(c, apperror.InvalidInputWithErr("invalid request body", err))
		return
	}

	w := &WeightReading{
		WeightKg:   req.WeightKg,
		BodyFatPct: req.BodyFatPct,
		BMI:        req.BMI,
		LeanMassKg: req.LeanMassKg,
		Date:       req.Date,
		Source:     req.Source,
	}

	if err := h.svc.CreateWeight(c.Request.Context(), w); err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

// ListVitals godoc
// @Summary      List vital readings
// @Description  Returns vital sign readings (heart rate, resting HR, blood pressure, etc.) for a metric across an optional date range.
// @Tags         apple-health
// @Produce      json
// @Security     BearerAuth
// @Param        metric  query  string  false  "Metric name (e.g. heart_rate, resting_heart_rate, blood_pressure_systolic)"
// @Param        from    query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to      query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Success      200  {object}  object{data=[]applehealth.Vital}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /apple-health/vitals [get]
// @ID           applehealth_list_vitals
func (h *Handler) ListVitals(c *gin.Context) {
	vitals, err := h.svc.ListVitals(c.Request.Context(),
		c.Query("metric"), c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vitals})
}

// IngestHealthData godoc
// @Summary      Bulk-ingest Apple Health data
// @Description  Accepts a HealthKit export payload (typically from a Shortcuts/Auto Export script) and persists weight + vitals records. Use for backfilling or scheduled pushes from a phone.
// @Tags         apple-health
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  applehealth.IngestPayload  true  "HealthKit ingest payload"
// @Success      200  {object}  object{status=string,records_added=integer}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /apple-health/ingest [post]
// @ID           applehealth_ingest
func (h *Handler) IngestHealthData(c *gin.Context) {
	var payload IngestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.RespondGin(c, apperror.InvalidInputWithErr("invalid JSON payload", err))
		return
	}

	added, err := Ingest(c.Request.Context(), h.repo, &payload)
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "records_added": added})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/apple-health")
	g.GET("/weight", h.ListWeight)
	g.POST("/weight", h.CreateWeight)
	g.GET("/vitals", h.ListVitals)
	g.POST("/ingest", h.IngestHealthData)
}
