package strava

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

// ListActivities godoc
// @Summary      List Strava activities
// @Description  Returns Strava activities (runs, rides, swims) within an optional date range and activity type filter. Use this for questions about workouts, distance, pace, or cardio output.
// @Tags         strava
// @Produce      json
// @Security     BearerAuth
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Param        type  query  string  false  "Activity type filter (e.g. Run, Ride, Swim, Walk)"
// @Success      200  {object}  object{data=[]strava.Activity}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /strava/activities [get]
// @ID           strava_list_activities
func (h *Handler) ListActivities(c *gin.Context) {
	activities, err := h.svc.List(c.Request.Context(),
		c.Query("from"), c.Query("to"), c.Query("type"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": activities})
}

// GetActivity godoc
// @Summary      Get a single Strava activity by id
// @Description  Returns full detail for one Strava activity including pace, heart rate, and elevation if available.
// @Tags         strava
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  integer  true  "Strava activity numeric id"
// @Success      200  {object}  object{data=strava.Activity}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      404  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /strava/activities/{id} [get]
// @ID           strava_get_activity
func (h *Handler) GetActivity(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apperror.RespondGin(c, apperror.InvalidInput("invalid activity id"))
		return
	}
	activity, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": activity})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/strava")
	g.GET("/activities", h.ListActivities)
	g.GET("/activities/:id", h.GetActivity)
}
