package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"user-service/internal/apperror"
	"user-service/internal/middleware"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type FavoritesHandler struct {
	favoritesService *service.FavoritesService
}

func NewFavoritesHandler(favoritesService *service.FavoritesService) *FavoritesHandler {
	return &FavoritesHandler{favoritesService: favoritesService}
}

// ListFavoriteVenues godoc
// @Summary      Список избранных площадок
// @Description  Возвращает список площадок, добавленных создателем в избранное
// @Tags         favorites
// @Produce      json
// @Success      200 {array} models.Venue
// @Failure      401 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/favorites/venues [get]
func (h *FavoritesHandler) ListFavoriteVenues(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	venues, err := h.favoritesService.ListFavoriteVenues(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch favorites"))
		return
	}

	c.JSON(http.StatusOK, venues)
}

// AddFavoriteVenue godoc
// @Summary      Добавить площадку в избранное
// @Description  Добавляет площадку в список избранных создателя
// @Tags         favorites
// @Produce      json
// @Param        user_id path int true "User ID площадки"
// @Success      200 {object} map[string]string
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/favorites/venues/{user_id} [put]
func (h *FavoritesHandler) AddFavoriteVenue(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	venueUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_PARAM", "Invalid user_id"))
		return
	}

	if err := h.favoritesService.AddFavoriteVenue(userID, venueUserID); err != nil {
		if errors.Is(err, service.ErrVenueNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("VENUE_NOT_FOUND", "Venue not found"))
			return
		}
		if errors.Is(err, service.ErrAlreadyFavorited) {
			c.JSON(http.StatusConflict, apperror.One("ALREADY_FAVORITED", "Venue is already in favorites"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to add favorite"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to favorites"})
}

// RemoveFavoriteVenue godoc
// @Summary      Убрать площадку из избранного
// @Description  Удаляет площадку из списка избранных создателя
// @Tags         favorites
// @Produce      json
// @Param        user_id path int true "User ID площадки"
// @Success      200 {object} map[string]string
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/favorites/venues/{user_id} [delete]
func (h *FavoritesHandler) RemoveFavoriteVenue(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	venueUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_PARAM", "Invalid user_id"))
		return
	}

	if err := h.favoritesService.RemoveFavoriteVenue(userID, venueUserID); err != nil {
		if errors.Is(err, service.ErrFavoriteNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("FAVORITE_NOT_FOUND", "Venue is not in favorites"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to remove favorite"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from favorites"})
}
