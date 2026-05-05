package hevy

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

// ListWorkouts godoc
// @Summary      List Hevy gym workouts
// @Description  Returns Hevy gym workouts (each with its sets) over an optional date range. Use this for strength training questions: volume, sets, reps, exercises lifted.
// @Tags         hevy
// @Produce      json
// @Security     BearerAuth
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive"
// @Success      200  {object}  object{data=[]hevy.Workout}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /hevy/workouts [get]
// @ID           hevy_list_workouts
func (h *Handler) ListWorkouts(c *gin.Context) {
	workouts, err := h.svc.List(c.Request.Context(), c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": workouts})
}

// GetWorkout godoc
// @Summary      Get one Hevy workout by id
// @Description  Returns a single gym workout with its full set list.
// @Tags         hevy
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  integer  true  "Hevy workout numeric id"
// @Success      200  {object}  object{data=hevy.Workout}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      404  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /hevy/workouts/{id} [get]
// @ID           hevy_get_workout
func (h *Handler) GetWorkout(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apperror.RespondGin(c, apperror.InvalidInput("invalid workout id"))
		return
	}
	workout, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": workout})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/hevy")
	g.GET("/workouts", h.ListWorkouts)
	g.GET("/workouts/:id", h.GetWorkout)
}
