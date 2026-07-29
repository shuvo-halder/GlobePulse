package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/global-news/analytics-service/internal/domain"
)

type AnalyticsHandler struct {
	AnalyticsUseCase domain.AnalyticsUseCase
}

func NewAnalyticsHandler(r *gin.Engine, us domain.AnalyticsUseCase) {
	handler := &AnalyticsHandler{
		AnalyticsUseCase: us,
	}

	api := r.Group("/api/v1/analytics")
	{
		api.GET("/country/:code", handler.GetCountryMetrics)
		api.GET("/global", handler.GetGlobalMetrics)
		api.GET("/heatmap", handler.GetHeatmapData)
	}
}

func (h *AnalyticsHandler) GetCountryMetrics(c *gin.Context) {
	code := c.Param("code")
	startStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date"})
		return
	}

	metrics, err := h.AnalyticsUseCase.GetCountryMetrics(c.Request.Context(), code, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}

func (h *AnalyticsHandler) GetGlobalMetrics(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date"})
		return
	}

	metrics, err := h.AnalyticsUseCase.GetGlobalMetrics(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}

func (h *AnalyticsHandler) GetHeatmapData(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date"})
		return
	}

	heatmap, err := h.AnalyticsUseCase.GetHeatmapData(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch heatmap data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": heatmap})
}
