package summary

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

// DailySummary godoc
// @Summary      Get daily cross-source summary
// @Description  Aggregates Oura, Strava, Hevy, InBody, Apple Health, and meals into a single daily summary. Use this as the default starting point when the user asks "how was my day" or "summarize my health on date X".
// @Tags         summary
// @Produce      json
// @Security     BearerAuth
// @Param        date  query  string  false  "Exact day (YYYY-MM-DD)"
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive — ignored if date is set"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive — ignored if date is set"
// @Success      200  {object}  object{data=[]summary.DailySummary}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /summary/daily [get]
// @ID           summary_daily
func (h *Handler) DailySummary(c *gin.Context) {
	summaries, err := h.svc.DailySummary(c.Request.Context(),
		c.Query("from"), c.Query("to"), c.Query("date"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summaries})
}

// WeeklyTrends godoc
// @Summary      Get weekly trend aggregates
// @Description  Returns weekly aggregates (averages and totals across sources). Use this when the user asks for week-over-week trends.
// @Tags         summary
// @Produce      json
// @Security     BearerAuth
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Success      200  {object}  object{data=[]summary.WeeklyTrend}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /summary/weekly [get]
// @ID           summary_weekly
func (h *Handler) WeeklyTrends(c *gin.Context) {
	trends, err := h.svc.WeeklyTrends(c.Request.Context(),
		c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trends})
}

// Report godoc
// @Summary      Get a multi-day health report
// @Description  Returns a rolling-window report ending at `date` covering the past `lookback` days (default 30). Use for "give me a report" requests.
// @Tags         summary
// @Produce      json
// @Security     BearerAuth
// @Param        date      query  string   false  "End date of the report (YYYY-MM-DD); defaults to today"
// @Param        lookback  query  integer  false  "Number of days to look back from date (default 30)"
// @Success      200  {object}  object{data=summary.ReportData}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /summary/report [get]
// @ID           summary_report
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
		apperror.RespondGin(c, err)
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
