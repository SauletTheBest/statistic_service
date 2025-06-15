package handler

import (
	"net/http"

	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)
// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	service  *service.AuthService
	validate *validator.Validate
	logger   *logrus.Logger
}
// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(s *service.AuthService, logger *logrus.Logger) *AuthHandler {
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
// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with the provided email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body authRequest true "User registration details"
// @Success 201 {object} map[string]string "status: success, message: user registered successfully"
// @Failure 400 {object} map[string]string "error: invalid request format or validation failed"
// @Failure 409 {object} map[string]string "error: user already exists"
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.logger.WithError(err).Warn("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}
	if err := h.service.Register(req.Email, req.Password); err != nil {
		h.logger.WithError(err).Warn("Registration failed")
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	h.logger.Info("User registered successfully")
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "user registered successfully"})
}
// Login godoc
// @Summary User login
// @Description Authenticates a user and returns access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body authRequest true "User login credentials"
// @Success 200 {object} map[string]string "access_token: JWT token, refresh_token: refresh token"
// @Failure 400 {object} map[string]string "error: invalid request format or validation failed"
// @Failure 401 {object} map[string]string "error: invalid email or password"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.logger.WithError(err).Warn("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}
	accessToken, refreshToken, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		h.logger.WithError(err).Warn("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	h.logger.Info("User logged in successfully")
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}
// Refresh godoc
// @Summary Refresh access token
// @Description Generates a new access token using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh token"
// @Success 200 {object} map[string]string "access_token: new JWT token"
// @Failure 400 {object} map[string]string "error: invalid request format or validation failed"
// @Failure 401 {object} map[string]string "error: invalid or expired refresh token"
// @Router /refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.logger.WithError(err).Warn("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}
	accessToken, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	h.logger.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
// GetProfile godoc
// @Summary Get user profile
// @Description Retrieves the authenticated user's profile information
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "id: user ID, email: user email"
// @Failure 401 {object} map[string]string "error: User not authenticated"
// @Failure 404 {object} map[string]string "error: User not found"
// @Router /me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
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
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "email": user.Email})
}
