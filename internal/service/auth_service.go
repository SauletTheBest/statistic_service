package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"statistic_service/pkg/jwt"
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, secret string) *AuthService {
	return &AuthService{repo, secret}
}

func (s *AuthService) Register(email, password string) (string, error) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &model.User{Email: email, PasswordHash: string(hashed)}

	if err := s.userRepo.Create(user); err != nil {
		return "", err
	}
	return jwt.GenerateToken(user.ID, s.jwtSecret)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return jwt.GenerateToken(user.ID, s.jwtSecret)
}

func (s *AuthService) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}
