package inbody

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

func (h *Handler) ListScans(c *gin.Context) {
	limit := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	scans, err := h.svc.List(c.Request.Context(), c.Query("from"), c.Query("to"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scans})
}

func (h *Handler) LatestScan(c *gin.Context) {
	scan, err := h.svc.Latest(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if scan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no body composition scans found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scan})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/inbody")
	g.GET("/scans", h.ListScans)
	g.GET("/scans/latest", h.LatestScan)
}
