package service

import (
	"errors"
	"regexp"

	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"statistic_service/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, secret string) *AuthService {
	return &AuthService{repo, secret}
}

func (s *AuthService) Register(email, password string) error {
	// Проверка на существующего пользователя
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return errors.New("user already exists")
	}

	// Дополнительная проверка сложности пароля
	if !isPasswordComplex(password) {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hashed),
	}

	return s.userRepo.Create(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	return jwt.GenerateToken(user.ID, s.jwtSecret)
}

func (s *AuthService) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// Проверка сложности пароля
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