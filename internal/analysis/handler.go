package analysis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-marty/health-connect/internal/apperror"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Analyze godoc
// @Summary      Generate AI health insights for a day
// @Description  Runs a Claude analysis over the daily summary and returns 3–4 specific bullet-point insights. Defaults to yesterday if `date` is not provided. Use this when the user wants commentary or interpretation rather than raw data.
// @Tags         analysis
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  analysis.AnalysisRequest  false  "Analysis request — date is optional and defaults to yesterday"
// @Success      200  {object}  object{data=analysis.AnalysisResponse}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /analysis [post]
// @ID           analysis_run
func (h *Handler) Analyze(c *gin.Context) {
	var req AnalysisRequest
	// Empty body is allowed — service defaults to yesterday.
	_ = c.ShouldBindJSON(&req)

	result, err := h.svc.Analyze(c.Request.Context(), req.Date)
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/analysis", h.Analyze)
}
