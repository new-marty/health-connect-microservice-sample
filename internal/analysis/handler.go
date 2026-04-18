package analysis

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Analyze(c *gin.Context) {
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body — default to yesterday
		req = AnalysisRequest{}
	}

	result, err := h.svc.Analyze(c.Request.Context(), req.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/analysis", h.Analyze)
}
