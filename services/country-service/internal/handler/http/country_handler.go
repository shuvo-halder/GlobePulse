package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/global-news/country-service/internal/domain"
)

type CountryHandler struct {
	CountryUseCase domain.CountryUseCase
}

func NewCountryHandler(r *gin.Engine, us domain.CountryUseCase) {
	handler := &CountryHandler{
		CountryUseCase: us,
	}

	api := r.Group("/api/v1/countries")
	{
		api.GET("", handler.List)
		api.GET("/:code", handler.GetDetails)
		api.GET("/rankings/:criteria", handler.Rankings)
	}
}

func (h *CountryHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := domain.CountryFilter{
		Query:  c.Query("q"),
		Region: c.Query("region"),
		SortBy: c.Query("sort_by"),
		Limit:  limit,
		Offset: offset,
	}

	countries, err := h.CountryUseCase.GetCountries(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch countries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": countries, "meta": gin.H{"limit": limit, "offset": offset}})
}

func (h *CountryHandler) GetDetails(c *gin.Context) {
	code := c.Param("code")

	country, err := h.CountryUseCase.GetCountryDetails(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": country})
}

func (h *CountryHandler) Rankings(c *gin.Context) {
	criteria := c.Param("criteria")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	countries, err := h.CountryUseCase.GetRankings(c.Request.Context(), criteria, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rankings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": countries})
}
