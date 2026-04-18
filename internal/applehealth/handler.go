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

func (h *Handler) ListWeight(c *gin.Context) {
	readings, err := h.svc.ListWeight(c.Request.Context(), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": readings})
}

type createWeightRequest struct {
	WeightKg   float64  `json:"weight_kg" binding:"required"`
	BodyFatPct *float64 `json:"body_fat_pct"`
	BMI        *float64 `json:"bmi"`
	LeanMassKg *float64 `json:"lean_mass_kg"`
	Date       string   `json:"date"`
	Source     string   `json:"source"`
}

func (h *Handler) CreateWeight(c *gin.Context) {
	var req createWeightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
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
		if apperror.IsInvalidInput(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *Handler) ListVitals(c *gin.Context) {
	vitals, err := h.svc.ListVitals(c.Request.Context(),
		c.Query("metric"), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vitals})
}

func (h *Handler) IngestHealthData(c *gin.Context) {
	var payload IngestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	added, err := Ingest(c.Request.Context(), h.repo, &payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
