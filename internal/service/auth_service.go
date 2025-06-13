package service

import (
	"errors"
	"regexp"
	"time"

	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"statistic_service/pkg/jwt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	logger    *logrus.Logger
}

func NewAuthService(repo repository.UserRepository, secret string, logger *logrus.Logger) *AuthService {
	return &AuthService{repo, secret, logger}
}

func (s *AuthService) Register(email, password string) error {
	s.logger.WithFields(logrus.Fields{
		"email": email,
	}).Info("Attempting to register user")

	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		s.logger.Warn("User already exists")
		return errors.New("user already exists")
	}

	if !isPasswordComplex(password) {
		s.logger.Warn("Password does not meet complexity requirements")
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash password")
		return err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hashed),
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.WithError(err).Error("Failed to create user")
		return err
	}

	s.logger.Info("User registered successfully")
	return nil
}

func (s *AuthService) Login(email, password string) (string, string, error) {
	s.logger.WithFields(logrus.Fields{
		"email": email,
	}).Info("Attempting to login user")

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		s.logger.Warn("Invalid email or password")
		return "", "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Invalid email or password")
		return "", "", errors.New("invalid email or password")
	}

	accessToken, err := jwt.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate access token")
		return "", "", err
	}

	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate refresh token")
		return "", "", err
	}

	refreshTokenModel := &model.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.userRepo.CreateRefreshToken(refreshTokenModel); err != nil {
		s.logger.WithError(err).Error("Failed to save refresh token")
		return "", "", err
	}

	s.logger.Info("User logged in successfully")
	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	s.logger.Info("Attempting to refresh token")

	token, err := s.userRepo.GetRefreshToken(refreshToken)
	if err != nil {
		s.logger.WithError(err).Warn("Invalid refresh token")
		return "", errors.New("invalid refresh token")
	}

	if time.Now().After(token.ExpiresAt) {
		s.logger.Warn("Refresh token expired")
		return "", errors.New("refresh token expired")
	}

	user, err := s.userRepo.GetByID(token.UserID)
	if err != nil {
		s.logger.WithError(err).Error("User not found for refresh token")
		return "", errors.New("user not found")
	}

	accessToken, err := jwt.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate new access token")
		return "", err
	}

	s.logger.Info("Token refreshed successfully")
	return accessToken, nil
}

func (s *AuthService) GetUserByID(id string) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user by ID")
		return nil, err
	}
	return user, nil
}

func isPasswordComplex(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	return hasUpper && hasLower && hasNumber && hasSpecial
}
