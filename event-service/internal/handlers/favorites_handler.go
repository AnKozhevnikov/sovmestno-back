package handlers

import (
	"errors"
	"event-service/internal/apperror"
	"event-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FavoritesHandler struct {
	favoritesService *service.FavoritesService
}

func NewFavoritesHandler(favoritesService *service.FavoritesService) *FavoritesHandler {
	return &FavoritesHandler{favoritesService: favoritesService}
}

// ListFavoriteEvents godoc
// @Summary      Список избранных мероприятий
// @Description  Возвращает список мероприятий, добавленных площадкой в избранное
// @Tags         favorites
// @Produce      json
// @Success      200 {array} models.Event
// @Failure      401 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /events/favorites [get]
func (h *FavoritesHandler) ListFavoriteEvents(c *gin.Context) {
	venueUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	events, err := h.favoritesService.ListFavoriteEvents(venueUserID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch favorites"))
		return
	}

	c.JSON(http.StatusOK, events)
}

// AddFavoriteEvent godoc
// @Summary      Добавить мероприятие в избранное
// @Description  Добавляет мероприятие в список избранных площадки
// @Tags         favorites
// @Produce      json
// @Param        id path int true "ID мероприятия"
// @Success      200 {object} map[string]string
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /events/favorites/{id} [put]
func (h *FavoritesHandler) AddFavoriteEvent(c *gin.Context) {
	venueUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_PARAM", "Invalid event id"))
		return
	}

	if err := h.favoritesService.AddFavoriteEvent(venueUserID.(int), eventID); err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("EVENT_NOT_FOUND", "Event not found"))
			return
		}
		if errors.Is(err, service.ErrAlreadyFavorited) {
			c.JSON(http.StatusConflict, apperror.One("ALREADY_FAVORITED", "Event is already in favorites"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to add favorite"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to favorites"})
}

// RemoveFavoriteEvent godoc
// @Summary      Убрать мероприятие из избранного
// @Description  Удаляет мероприятие из списка избранных площадки
// @Tags         favorites
// @Produce      json
// @Param        id path int true "ID мероприятия"
// @Success      200 {object} map[string]string
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /events/favorites/{id} [delete]
func (h *FavoritesHandler) RemoveFavoriteEvent(c *gin.Context) {
	venueUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_PARAM", "Invalid event id"))
		return
	}

	if err := h.favoritesService.RemoveFavoriteEvent(venueUserID.(int), eventID); err != nil {
		if errors.Is(err, service.ErrFavoriteNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("FAVORITE_NOT_FOUND", "Event is not in favorites"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to remove favorite"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from favorites"})
}
