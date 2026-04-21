package service

import (
	"errors"
	"event-service/internal/models"
	"event-service/internal/repository"

	"gorm.io/gorm"
)

type FavoritesService struct {
	repo repository.EventRepositoryInterface
}

func NewFavoritesService(repo repository.EventRepositoryInterface) *FavoritesService {
	return &FavoritesService{repo: repo}
}

func (s *FavoritesService) AddFavoriteEvent(venueUserID, eventID int) error {
	_, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return ErrEventNotFound
	}

	alreadyExisted, err := s.repo.AddVenueFavoriteEvent(venueUserID, eventID)
	if err != nil {
		return err
	}
	if alreadyExisted {
		return ErrAlreadyFavorited
	}
	return nil
}

func (s *FavoritesService) RemoveFavoriteEvent(venueUserID, eventID int) error {
	err := s.repo.RemoveVenueFavoriteEvent(venueUserID, eventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrFavoriteNotFound
	}
	return err
}

func (s *FavoritesService) ListFavoriteEvents(venueUserID int) ([]models.Event, error) {
	return s.repo.ListVenueFavoriteEvents(venueUserID)
}
