package handlers

import (
	"net/http"
	"strconv"
	"time"
	"user-service/internal/apperror"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
)

type PublicCreatorResponse struct {
	ID          int                   `json:"id"`
	UserID      int                   `json:"user_id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	PhotoID     *string               `json:"photo_id,omitempty"`
	TgChannel   string                `json:"tg_channel_link,omitempty"`
	VkLink      string                `json:"vk_link,omitempty"`
	TiktokLink  string                `json:"tiktok_link,omitempty"`
	YoutubeLink string                `json:"youtube_link,omitempty"`
	DzenLink    string                `json:"dzen_link,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Photo       *models.Image         `json:"photo,omitempty"`
	Photos      []models.CreatorPhoto `json:"photos,omitempty"`
}

type PublicVenueResponse struct {
	ID            int                 `json:"id"`
	UserID        int                 `json:"user_id"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	StreetAddress string              `json:"street_address,omitempty"`
	CityID        *int                `json:"city_id,omitempty"`
	OpeningHours  string              `json:"opening_hours,omitempty"`
	Capacity      int                 `json:"capacity,omitempty"`
	LogoID        *string             `json:"logo_id,omitempty"`
	CoverPhotoID  *string             `json:"cover_photo_id,omitempty"`
	TgChannel     string              `json:"tg_channel_link,omitempty"`
	VkLink        string              `json:"vk_link,omitempty"`
	TiktokLink    string              `json:"tiktok_link,omitempty"`
	YoutubeLink   string              `json:"youtube_link,omitempty"`
	DzenLink      string              `json:"dzen_link,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Logo          *models.Image       `json:"logo,omitempty"`
	CoverPhoto    *models.Image       `json:"cover_photo,omitempty"`
	Photos        []models.VenuePhoto `json:"photos,omitempty"`
}

func toPublicCreator(c *models.Creator) PublicCreatorResponse {
	return PublicCreatorResponse{
		ID:          c.ID,
		UserID:      c.UserID,
		Name:        c.Name,
		Description: c.Description,
		PhotoID:     c.PhotoID,
		TgChannel:   c.TgChannel,
		VkLink:      c.VkLink,
		TiktokLink:  c.TiktokLink,
		YoutubeLink: c.YoutubeLink,
		DzenLink:    c.DzenLink,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Photo:       c.Photo,
		Photos:      c.Photos,
	}
}

func toPublicVenue(v *models.Venue) PublicVenueResponse {
	return PublicVenueResponse{
		ID:            v.ID,
		UserID:        v.UserID,
		Name:          v.Name,
		Description:   v.Description,
		StreetAddress: v.StreetAddress,
		CityID:        v.CityID,
		OpeningHours:  v.OpeningHours,
		Capacity:      v.Capacity,
		LogoID:        v.LogoID,
		CoverPhotoID:  v.CoverPhotoID,
		TgChannel:     v.TgChannel,
		VkLink:        v.VkLink,
		TiktokLink:    v.TiktokLink,
		YoutubeLink:   v.YoutubeLink,
		DzenLink:      v.DzenLink,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
		Logo:          v.Logo,
		CoverPhoto:    v.CoverPhoto,
		Photos:        v.Photos,
	}
}

// GetPublicCreator godoc
// @Summary      Публичный профиль создателя
// @Description  Возвращает профиль создателя без личных контактов (phone, email, личный TG)
// @Tags         public
// @Produce      json
// @Param        user_id path int true "User ID"
// @Success      200 {object} PublicCreatorResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /public/creators/{user_id} [get]
func (h *UserHandler) GetPublicCreator(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid user ID"))
		return
	}

	creator, err := h.userService.GetCreatorByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, apperror.One("CREATOR_NOT_FOUND", "Creator not found"))
		return
	}

	c.JSON(http.StatusOK, toPublicCreator(creator))
}

// ListPublicCreators godoc
// @Summary      Публичный список создателей
// @Description  Возвращает список создателей без личных контактов
// @Tags         public
// @Produce      json
// @Param        limit  query int false "Количество элементов (по умолчанию 20)"
// @Param        offset query int false "Смещение (по умолчанию 0)"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /public/creators [get]
func (h *UserHandler) ListPublicCreators(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	creators, err := h.userService.ListCreators(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch creators"))
		return
	}

	public := make([]PublicCreatorResponse, len(creators))
	for i := range creators {
		public[i] = toPublicCreator(&creators[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"creators": public,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetPublicVenue godoc
// @Summary      Публичный профиль площадки
// @Description  Возвращает профиль площадки без личных контактов (phone, email, личный TG)
// @Tags         public
// @Produce      json
// @Param        user_id path int true "User ID"
// @Success      200 {object} PublicVenueResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /public/venues/{user_id} [get]
func (h *UserHandler) GetPublicVenue(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid user ID"))
		return
	}

	venue, err := h.userService.GetVenueByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, apperror.One("VENUE_NOT_FOUND", "Venue not found"))
		return
	}

	c.JSON(http.StatusOK, toPublicVenue(venue))
}

// ListPublicVenues godoc
// @Summary      Публичный список площадок
// @Description  Возвращает список площадок без личных контактов
// @Tags         public
// @Produce      json
// @Param        limit  query int false "Количество элементов (по умолчанию 20)"
// @Param        offset query int false "Смещение (по умолчанию 0)"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /public/venues [get]
func (h *UserHandler) ListPublicVenues(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	venues, err := h.userService.ListVenues(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch venues"))
		return
	}

	public := make([]PublicVenueResponse, len(venues))
	for i := range venues {
		public[i] = toPublicVenue(&venues[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"venues": public,
		"limit":  limit,
		"offset": offset,
	})
}

