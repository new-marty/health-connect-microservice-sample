package sync

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      *Repository
	scheduler *Scheduler
}

func NewHandler(repo *Repository, scheduler *Scheduler) *Handler {
	return &Handler{repo: repo, scheduler: scheduler}
}

func (h *Handler) GetStatus(c *gin.Context) {
	statuses, err := h.repo.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": statuses})
}

func (h *Handler) TriggerSync(c *gin.Context) {
	source := c.Param("source")

	added, err := h.scheduler.TriggerSync(c.Request.Context(), source)
	if err != nil {
		if _, ok := err.(*unknownSourceError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"records_added": added,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"source":        source,
		"records_added": added,
	})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/sync")
	g.GET("/status", h.GetStatus)
	g.POST("/:source", h.TriggerSync)
}
