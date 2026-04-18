package summary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) DailySummary(c *gin.Context) {
	summaries, err := h.svc.DailySummary(c.Request.Context(),
		c.Query("from"), c.Query("to"), c.Query("date"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summaries})
}

func (h *Handler) WeeklyTrends(c *gin.Context) {
	trends, err := h.svc.WeeklyTrends(c.Request.Context(),
		c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trends})
}

func (h *Handler) Report(c *gin.Context) {
	date := c.Query("date")
	lookback := 30
	if l := c.Query("lookback"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			lookback = n
		}
	}

	report, err := h.svc.Report(c.Request.Context(), date, lookback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/summary")
	g.GET("/daily", h.DailySummary)
	g.GET("/weekly", h.WeeklyTrends)
	g.GET("/report", h.Report)
}
