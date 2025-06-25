package service

import (
	"statistic_service/internal/model"
	"statistic_service/internal/repository"

	"github.com/sirupsen/logrus"
)

// UserService определяет интерфейс для операций с пользователями.
type UserService interface {
	GetUserByID(userID string) (*model.User, error)
	// Добавь здесь другие методы, связанные с пользователями, если понадобятся
}

// UserServiceImpl реализует UserService.
type UserServiceImpl struct {
	userRepo repository.UserRepository
	logger   *logrus.Logger
}

// NewUserService создает новый экземпляр UserServiceImpl.
func NewUserService(userRepo repository.UserRepository, logger *logrus.Logger) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUserByID получает пользователя по его ID.
func (s *UserServiceImpl) GetUserByID(userID string) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user by ID from repository")
		return nil, err
	}
	return user, nil
}
