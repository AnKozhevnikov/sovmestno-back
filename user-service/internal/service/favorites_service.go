package service

import (
	"errors"
	"user-service/internal/models"
	"user-service/internal/repository"

	"gorm.io/gorm"
)

type FavoritesService struct {
	repo *repository.UserRepository
}

func NewFavoritesService(repo *repository.UserRepository) *FavoritesService {
	return &FavoritesService{repo: repo}
}

func (s *FavoritesService) AddFavoriteVenue(creatorUserID, venueUserID int) error {
	_, err := s.repo.GetVenueByUserID(venueUserID)
	if err != nil {
		return ErrVenueNotFound
	}

	alreadyExisted, err := s.repo.AddCreatorFavoriteVenue(creatorUserID, venueUserID)
	if err != nil {
		return err
	}
	if alreadyExisted {
		return ErrAlreadyFavorited
	}
	return nil
}

func (s *FavoritesService) RemoveFavoriteVenue(creatorUserID, venueUserID int) error {
	err := s.repo.RemoveCreatorFavoriteVenue(creatorUserID, venueUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrFavoriteNotFound
	}
	return err
}

func (s *FavoritesService) ListFavoriteVenues(creatorUserID int) ([]models.Venue, error) {
	return s.repo.ListCreatorFavoriteVenues(creatorUserID)
}
