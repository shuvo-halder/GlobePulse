package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/global-news/auth-service/internal/domain"
	"github.com/global-news/auth-service/internal/handler/http/dto"
)

type AuthHandler struct {
	AuthUseCase domain.AuthUseCase
}

func NewAuthHandler(r *gin.Engine, us domain.AuthUseCase, authMiddleware gin.HandlerFunc) {
	handler := &AuthHandler{
		AuthUseCase: us,
	}

	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/reset-password", handler.ResetPassword)

		protected := api.Group("/")
		protected.Use(authMiddleware)
		{
			protected.POST("/logout", handler.Logout)
			protected.GET("/profile", handler.Profile)
		}
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.AuthUseCase.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, sessionID, err := h.AuthUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"session_id": sessionID,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, _ := c.Get("session_id")

	err := h.AuthUseCase.Logout(c.Request.Context(), sessionID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.AuthUseCase.ResetPassword(c.Request.Context(), req.Email, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.AuthUseCase.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}
