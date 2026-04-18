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

func (h *Handler) ListMeals(c *gin.Context) {
	meals, err := h.svc.List(c.Request.Context(),
		c.Query("date"), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": meals})
}

type createMealRequest struct {
	Date        string  `json:"date" binding:"required"`
	Meal        string  `json:"meal"`
	Description string  `json:"description" binding:"required"`
	Calories    int     `json:"calories"`
	ProteinG    float64 `json:"protein_g"`
	FatG        float64 `json:"fat_g"`
	CarbsG      float64 `json:"carbs_g"`
	Source      string  `json:"source"`
}

func (h *Handler) CreateMeal(c *gin.Context) {
	var req createMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
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
		if apperror.IsInvalidInput(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) DeleteMeal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meal id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if apperror.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/meals", h.ListMeals)
	rg.POST("/meals", h.CreateMeal)
	rg.DELETE("/meals/:id", h.DeleteMeal)
}
