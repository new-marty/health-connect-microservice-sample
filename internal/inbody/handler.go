package inbody

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-marty/health-connect/internal/apperror"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListScans godoc
// @Summary      List InBody body composition scans
// @Description  Returns InBody scans (weight, body fat %, muscle mass, etc.) over an optional date range. Use for body composition trends.
// @Tags         inbody
// @Produce      json
// @Security     BearerAuth
// @Param        from   query  string   false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to     query  string   false  "End date (YYYY-MM-DD), inclusive"
// @Param        limit  query  integer  false  "Maximum number of scans to return (newest first)"
// @Success      200  {object}  object{data=[]inbody.BodyCompScan}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /inbody/scans [get]
// @ID           inbody_list_scans
func (h *Handler) ListScans(c *gin.Context) {
	limit := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	scans, err := h.svc.List(c.Request.Context(), c.Query("from"), c.Query("to"), limit)
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scans})
}

// LatestScan godoc
// @Summary      Get the most recent InBody scan
// @Description  Returns the single most recent InBody body composition scan, or 404 if none exist.
// @Tags         inbody
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{data=inbody.BodyCompScan}
// @Failure      401  {object}  apperror.Response
// @Failure      404  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /inbody/scans/latest [get]
// @ID           inbody_get_latest_scan
func (h *Handler) LatestScan(c *gin.Context) {
	scan, err := h.svc.Latest(c.Request.Context())
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	if scan == nil {
		apperror.RespondGin(c, apperror.NotFound("inbody scan"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scan})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/inbody")
	g.GET("/scans", h.ListScans)
	g.GET("/scans/latest", h.LatestScan)
}
