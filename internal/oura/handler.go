package oura

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

func (h *Handler) ListSleep(c *gin.Context) {
	sessions, err := h.svc.ListSleep(c.Request.Context(),
		c.Query("from"), c.Query("to"), c.Query("type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *Handler) GetSleepByDay(c *gin.Context) {
	sessions, err := h.svc.GetSleepByDay(c.Request.Context(), c.Param("day"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(sessions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no sleep data for this day"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *Handler) ListScores(c *gin.Context) {
	scores, err := h.svc.ListScores(c.Request.Context(),
		c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scores})
}

func (h *Handler) GetScoreByDay(c *gin.Context) {
	score, err := h.svc.GetScoreByDay(c.Request.Context(), c.Param("day"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if score == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no scores for this day"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": score})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/oura")
	g.GET("/sleep", h.ListSleep)
	g.GET("/sleep/:day", h.GetSleepByDay)
	g.GET("/scores", h.ListScores)
	g.GET("/scores/:day", h.GetScoreByDay)
}
