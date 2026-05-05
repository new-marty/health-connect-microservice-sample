package meals

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

// ListMeals godoc
// @Summary      List logged meals
// @Description  Returns logged meals filtered by an exact date or a date range. Use for nutrition questions: calories, macros, what was eaten on a date.
// @Tags         meals
// @Produce      json
// @Security     BearerAuth
// @Param        date  query  string  false  "Exact day (YYYY-MM-DD)"
// @Param        from  query  string  false  "Start date (YYYY-MM-DD), inclusive — ignored if date is set"
// @Param        to    query  string  false  "End date (YYYY-MM-DD), inclusive — ignored if date is set"
// @Success      200  {object}  object{data=[]meals.Meal}
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /meals [get]
// @ID           meals_list
func (h *Handler) ListMeals(c *gin.Context) {
	meals, err := h.svc.List(c.Request.Context(),
		c.Query("date"), c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": meals})
}

// CreateMealRequest is the JSON body for logging a meal.
type CreateMealRequest struct {
	Date        string  `json:"date" binding:"required" example:"2026-05-04"`
	Meal        string  `json:"meal,omitempty" example:"lunch"`
	Description string  `json:"description" binding:"required" example:"chicken rice bowl with broccoli"`
	Calories    int     `json:"calories,omitempty" example:"650"`
	ProteinG    float64 `json:"protein_g,omitempty" example:"45"`
	FatG        float64 `json:"fat_g,omitempty" example:"18"`
	CarbsG      float64 `json:"carbs_g,omitempty" example:"70"`
	Source      string  `json:"source,omitempty" example:"manual"`
}

// CreateMeal godoc
// @Summary      Log a meal
// @Description  Creates a meal log entry. Use this when the user wants to record what they ate.
// @Tags         meals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  meals.CreateMealRequest  true  "Meal log entry"
// @Success      201  {object}  object{id=integer}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /meals [post]
// @ID           meals_create
func (h *Handler) CreateMeal(c *gin.Context) {
	var req CreateMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondGin(c, apperror.InvalidInputWithErr("invalid request body", err))
		return
	}

	m := &Meal{
		Date:        req.Date,
		Meal:        req.Meal,
		Description: req.Description,
		Calories:    req.Calories,
		ProteinG:    req.ProteinG,
		FatG:        req.FatG,
		CarbsG:      req.CarbsG,
		Source:      req.Source,
	}

	id, err := h.svc.Create(c.Request.Context(), m)
	if err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// DeleteMeal godoc
// @Summary      Delete a meal log entry
// @Description  Removes a meal entry by id.
// @Tags         meals
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  integer  true  "Meal id"
// @Success      200  {object}  object{status=string}
// @Failure      400  {object}  apperror.Response
// @Failure      401  {object}  apperror.Response
// @Failure      404  {object}  apperror.Response
// @Failure      500  {object}  apperror.Response
// @Router       /meals/{id} [delete]
// @ID           meals_delete
func (h *Handler) DeleteMeal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apperror.RespondGin(c, apperror.InvalidInput("invalid meal id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		apperror.RespondGin(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/meals", h.ListMeals)
	rg.POST("/meals", h.CreateMeal)
	rg.DELETE("/meals/:id", h.DeleteMeal)
}
