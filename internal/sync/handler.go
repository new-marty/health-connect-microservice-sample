package sync

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-marty/health-connect/internal/apperror"
)

type Handler struct {
	repo      *Repository
	scheduler *Scheduler
}

func NewHandler(repo *Repository, scheduler *Scheduler) *Handler {
	return &Handler{repo: repo, scheduler: scheduler}
}

// GetStatus godoc
// @Summary      Get sync status per source
// @Description  Returns last successful sync timestamp, last status, and total records added per source.
// @Tags         sync
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{data=[]sync.SyncStatus}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /sync/status [get]
// @ID           sync_status
func (h *Handler) GetStatus(c *gin.Context) {
	statuses, err := h.repo.GetStatus(c.Request.Context())
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": statuses})
}

// TriggerSync godoc
// @Summary      Trigger an on-demand sync
// @Description  Forces a sync for the named source. Pass "all" to sync every registered source. Use when the user wants their data refreshed before analysis.
// @Tags         sync
// @Produce      json
// @Security     BearerAuth
// @Param        source  path  string  true  "Source name: oura, strava, hevy, inbody, or all"
// @Success      200  {object}  object{status=string,source=string,records_added=integer}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /sync/{source} [post]
// @ID           sync_trigger
func (h *Handler) TriggerSync(c *gin.Context) {
	source := c.Param("source")

	added, err := h.scheduler.TriggerSync(c.Request.Context(), source)
	if err != nil {
		if _, ok := err.(*unknownSourceError); ok {
			apperror.RespondGin(c, apperror.InvalidInput(err.Error()))
			return
		}
		apperror.RespondGin(c, err)
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
