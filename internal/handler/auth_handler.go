package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"statistic_service/internal/service"
)

type AuthHandler struct {
	service  *service.AuthService
	validate *validator.Validate
	logger   *logrus.Logger
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &AuthHandler{
		service:  s,
		validate: validator.New(),
		logger:   logger,
	}
}

type authRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		h.logger.WithError(err).Warn("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors.Error()})
		return
	}

	err := h.service.Register(req.Email, req.Password)
	if err != nil {
		h.logger.WithError(err).Warn("Registration failed")
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User registered successfully")
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "user registered successfully",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		h.logger.WithError(err).Warn("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors.Error()})
		return
	}

	accessToken, refreshToken, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		h.logger.WithError(err).Warn("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User logged in successfully")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		h.logger.WithError(err).Warn("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors.Error()})
		return
	}

	accessToken, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.service.GetUserByID(userID.(string))
	if err != nil {
		h.logger.WithError(err).Warn("User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.logger.Info("User profile retrieved successfully")
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
	})
}