package service

import (
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NewsletterService struct {
	repo *repository.UserRepository
}

func NewNewsletterService(repo *repository.UserRepository) *NewsletterService {
	return &NewsletterService{repo: repo}
}

func (s *NewsletterService) Subscribe(email string) (*models.NewsletterSubscription, error) {
	existing, err := s.repo.GetNewsletterSubscriptionByEmail(email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadySubscribed
	}

	sub := &models.NewsletterSubscription{Email: email, UnsubscribeToken: uuid.New().String()}
	if err := s.repo.CreateNewsletterSubscription(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *NewsletterService) UnsubscribeByToken(token string) error {
	sub, err := s.repo.GetNewsletterSubscriptionByToken(token)
	if err != nil {
		return ErrInvalidUnsubscribeToken
	}
	return s.repo.DeleteNewsletterSubscription(sub.ID)
}

func (s *NewsletterService) ListSubscriptions() ([]models.NewsletterSubscription, error) {
	return s.repo.ListNewsletterSubscriptions()
}
