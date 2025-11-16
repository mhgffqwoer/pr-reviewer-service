package services

import (
	"github.com/mhgffqwoer/pr-service/internal/models"
)

type UserRepository interface {
	GetByID(userID string) (*models.User, error)
	Exists(userID string) bool
	Save(user *models.User) error
	GetReview(userID string) ([]*models.PullRequestShort, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SetActive(userID string, isActive bool) (*models.User, error) {
	if !s.repo.Exists(userID) {
		return nil, ErrNotFound
	}

	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	user.IsActive = isActive
	_ = s.repo.Save(user)

	return user, nil
}

func (s *UserService) GetReviewe(userID string) ([]*models.PullRequestShort, error) {
	if !s.repo.Exists(userID) {
		return nil, ErrNotFound
	}

	reviews, _ := s.repo.GetReview(userID)
	return reviews, nil
}
