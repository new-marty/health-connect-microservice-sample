package oura

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

// ListSleep godoc
// @Summary      List Oura sleep sessions
// @Description  Returns sleep sessions from the Oura Ring within an optional date range and type filter. Use this when the user asks how they slept, sleep duration, sleep stages, or HRV during sleep.
// @Tags         oura
// @Produce      json
// @Security     BearerAuth
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Param        type  query  string  false  "Sleep type filter: long, short, or nap"
// @Success      200  {object}  object{data=[]oura.SleepSession}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /oura/sleep [get]
// @ID           oura_list_sleep
func (h *Handler) ListSleep(c *gin.Context) {
	sessions, err := h.svc.ListSleep(c.Request.Context(),
		c.Query("from"), c.Query("to"), c.Query("type"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// GetSleepByDay godoc
// @Summary      Get Oura sleep for a specific day
// @Description  Returns sleep sessions recorded for a single calendar day. Use this when the user asks about a specific date's sleep.
// @Tags         oura
// @Produce      json
// @Security     BearerAuth
// @Param        day  path  string  true  "Calendar day (YYYY-MM-DD)"
// @Success      200  {object}  object{data=[]oura.SleepSession}
// @Failure      401  {object}  apperror.Response
// @Failure      404  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /oura/sleep/{day} [get]
// @ID           oura_get_sleep_by_day
func (h *Handler) GetSleepByDay(c *gin.Context) {
	sessions, err := h.svc.GetSleepByDay(c.Request.Context(), c.Param("day"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	if len(sessions) == 0 {
		apperror.RespondGin(c, apperror.NotFound("oura sleep for day"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// ListScores godoc
// @Summary      List Oura daily scores
// @Description  Returns daily Oura scores (sleep, readiness, activity, HRV) over a date range. Use this for trend questions or when the user asks for readiness/activity/sleep scores.
// @Tags         oura
// @Produce      json
// @Security     BearerAuth
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Success      200  {object}  object{data=[]oura.DailyScore}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /oura/scores [get]
// @ID           oura_list_scores
func (h *Handler) ListScores(c *gin.Context) {
	scores, err := h.svc.ListScores(c.Request.Context(),
		c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scores})
}

// GetScoreByDay godoc
// @Summary      Get Oura scores for a specific day
// @Description  Returns the daily Oura score (sleep, readiness, activity) for one calendar day.
// @Tags         oura
// @Produce      json
// @Security     BearerAuth
// @Param        day  path  string  true  "Calendar day (YYYY-MM-DD)"
// @Success      200  {object}  object{data=oura.DailyScore}
// @Failure      401  {object}  apperror.Response
// @Failure      404  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /oura/scores/{day} [get]
// @ID           oura_get_score_by_day
func (h *Handler) GetScoreByDay(c *gin.Context) {
	score, err := h.svc.GetScoreByDay(c.Request.Context(), c.Param("day"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	if score == nil {
		apperror.RespondGin(c, apperror.NotFound("oura score for day"))
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
