package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/global-news/news-service/internal/domain"
	"github.com/global-news/news-service/pkg/logger"
	types "github.com/global-news/shared-types"
	"go.uber.org/zap"
)

type NewsHandler struct {
	NewsUseCase domain.NewsUseCase
}

func NewNewsHandler(r *gin.Engine, us domain.NewsUseCase) {
	handler := &NewsHandler{
		NewsUseCase: us,
	}

	api := r.Group("/api/v1/news")
	{
		api.GET("", handler.FetchNews)
		api.GET("/:id", handler.GetByID)
		api.POST("", handler.Create)
	}
}

func (h *NewsHandler) FetchNews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := domain.NewsFilter{
		CountryCode: c.Query("country_code"),
		Topic:       c.Query("topic"),
		Query:       c.Query("query"),
		Limit:       limit,
		Offset:      offset,
	}

	news, err := h.NewsUseCase.GetNews(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": news, "meta": gin.H{"limit": limit, "offset": offset}})
}

func (h *NewsHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	article, err := h.NewsUseCase.GetArticleByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func (h *NewsHandler) Create(c *gin.Context) {
	var article types.NewsArticle
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.NewsUseCase.CreateArticle(c.Request.Context(), &article); err != nil {
		logger.Log.Error("Failed to create article", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": article})
}
