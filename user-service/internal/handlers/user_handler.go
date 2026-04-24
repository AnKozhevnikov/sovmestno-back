package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"user-service/internal/apperror"
	"user-service/internal/middleware"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService  *service.UserService
	imageService *service.ImageService
}

func NewUserHandler(userService *service.UserService, imageService *service.ImageService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		imageService: imageService,
	}
}

// Creator handlers

// CreateCreator godoc
// @Summary      Создать профиль создателя
// @Description  Создает профиль создателя мероприятий для текущего пользователя
// @Tags         creators
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.CreateCreatorRequest true "Данные профиля создателя"
// @Success      201 {object} models.Creator "Профиль создателя создан"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /users/creators [post]
func (h *UserHandler) CreateCreator(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	var req service.CreateCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	creator, err := h.userService.CreateCreator(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			c.JSON(http.StatusConflict, apperror.One("PROFILE_ALREADY_EXISTS", "Creator profile already exists for this user"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create creator profile"))
		return
	}

	c.JSON(http.StatusCreated, creator)
}

// GetCreator godoc
// @Summary      Получить профиль создателя
// @Description  Возвращает профиль создателя по user_id
// @Tags         creators
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      200 {object} models.Creator "Профиль создателя"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/creators/{user_id} [get]
func (h *UserHandler) GetCreator(c *gin.Context) {
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

	c.JSON(http.StatusOK, creator)
}

// GetMe godoc
// @Summary      Получить мой профиль
// @Description  Возвращает профиль текущего пользователя (создателя или площадки) в зависимости от роли
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Профиль пользователя с ролью и данными профиля"
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	profile, err := h.userService.GetMyProfile(userID)
	if err != nil {
		if errors.Is(err, service.ErrCreatorNotFound) || errors.Is(err, service.ErrVenueNotFound) || errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("PROFILE_NOT_FOUND", "Profile not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch profile"))
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateCreator godoc
// @Summary      Обновить профиль создателя
// @Description  Обновляет профиль создателя (только свой профиль)
// @Tags         creators
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Param        request body service.UpdateCreatorRequest true "Обновленные данные профиля"
// @Success      200 {object} models.Creator "Профиль обновлен"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/creators/{user_id} [put]
func (h *UserHandler) UpdateCreator(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid user ID"))
		return
	}

	var req service.UpdateCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	creator, err := h.userService.UpdateCreatorByUserID(targetUserID, currentUserID, &req)
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You can only edit your own profile"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("CREATOR_NOT_FOUND", "Creator not found"))
		return
	}

	c.JSON(http.StatusOK, creator)
}

// DeleteCreator godoc
// @Summary      Удалить профиль создателя
// @Description  Удаляет профиль создателя (только свой профиль)
// @Tags         creators
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      204 "Профиль удален"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Router       /users/creators/{user_id} [delete]
func (h *UserHandler) DeleteCreator(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid user ID"))
		return
	}

	if err := h.userService.DeleteCreatorByUserID(targetUserID, currentUserID); err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You can only delete your own profile"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("CREATOR_NOT_FOUND", "Creator not found"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListCreators godoc
// @Summary      Список создателей
// @Description  Возвращает список создателей мероприятий с пагинацией
// @Tags         creators
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Количество элементов (по умолчанию 20, максимум 100)"
// @Param        offset query int false "Смещение (по умолчанию 0)"
// @Success      200 {object} map[string]interface{} "Список создателей"
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /users/creators [get]
func (h *UserHandler) ListCreators(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	creators, err := h.userService.ListCreators(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch creators"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"creators": creators,
		"limit":    limit,
		"offset":   offset,
	})
}

// Venue handlers

// CreateVenue godoc
// @Summary      Создать профиль площадки
// @Description  Создает профиль площадки для текущего пользователя
// @Tags         venues
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.CreateVenueRequest true "Данные профиля площадки"
// @Success      201 {object} models.Venue "Профиль площадки создан"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /users/venues [post]
func (h *UserHandler) CreateVenue(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	var req service.CreateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	venue, err := h.userService.CreateVenue(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			c.JSON(http.StatusConflict, apperror.One("PROFILE_ALREADY_EXISTS", "Venue profile already exists for this user"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create venue profile"))
		return
	}

	c.JSON(http.StatusCreated, venue)
}

// GetVenue godoc
// @Summary      Получить профиль площадки
// @Description  Возвращает профиль площадки по user_id
// @Tags         venues
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      200 {object} models.Venue "Профиль площадки"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/venues/{user_id} [get]
func (h *UserHandler) GetVenue(c *gin.Context) {
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

	c.JSON(http.StatusOK, venue)
}

// ListVenues godoc
// @Summary      Список площадок
// @Description  Возвращает список площадок с пагинацией
// @Tags         venues
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Количество элементов (по умолчанию 20, максимум 100)"
// @Param        offset query int false "Смещение (по умолчанию 0)"
// @Success      200 {object} map[string]interface{} "Список площадок"
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /users/venues [get]
func (h *UserHandler) ListVenues(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	venues, err := h.userService.ListVenues(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch venues"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"venues": venues,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateVenue godoc
// @Summary      Обновить профиль площадки
// @Description  Обновляет профиль площадки (только свой профиль)
// @Tags         venues
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Param        request body service.UpdateVenueRequest true "Обновленные данные профиля"
// @Success      200 {object} models.Venue "Профиль обновлен"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/venues/{user_id} [put]
func (h *UserHandler) UpdateVenue(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid user ID"))
		return
	}

	var req service.UpdateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	venue, err := h.userService.UpdateVenueByUserID(targetUserID, currentUserID, &req)
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You can only edit your own profile"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("VENUE_NOT_FOUND", "Venue not found"))
		return
	}

	c.JSON(http.StatusOK, venue)
}

// DeleteVenue godoc
// @Summary      Удалить профиль площадки
// @Description  Удаляет профиль площадки (только свой профиль)
// @Tags         venues
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      204 "Профиль удален"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Router       /users/venues/{user_id} [delete]
func (h *UserHandler) DeleteVenue(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid user ID"))
		return
	}

	if err := h.userService.DeleteVenueByUserID(targetUserID, currentUserID); err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You can only delete your own profile"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("VENUE_NOT_FOUND", "Venue not found"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddCreatorPhoto godoc
// @Summary      Добавить фото создателя
// @Description  Привязывает загруженное изображение к галерее создателя
// @Tags         creators
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body map[string]string true "UUID изображения" example({"image_id": "550e8400-e29b-41d4-a716-446655440000"})
// @Success      201 {object} models.CreatorPhoto "Фото добавлено"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/creators/photos [post]
func (h *UserHandler) AddCreatorPhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	var req struct {
		ImageID string `json:"image_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	photo, err := h.userService.AddCreatorPhoto(userID, req.ImageID)
	if err != nil {
		if errors.Is(err, service.ErrCreatorNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("CREATOR_NOT_FOUND", "Creator profile not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to add photo"))
		return
	}

	c.JSON(http.StatusCreated, photo)
}

// DeleteCreatorPhoto godoc
// @Summary      Удалить фото создателя
// @Description  Удаляет фото из галереи создателя (только своё фото)
// @Tags         creators
// @Security     BearerAuth
// @Param        photo_id path int true "ID записи фото"
// @Success      204 "Фото удалено"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/creators/photos/{photo_id} [delete]
func (h *UserHandler) DeleteCreatorPhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	photoID, err := strconv.Atoi(c.Param("photo_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid photo ID"))
		return
	}

	if err := h.userService.DeleteCreatorPhoto(userID, photoID); err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You can only delete your own photos"))
			return
		}
		if errors.Is(err, service.ErrPhotoNotFound) || errors.Is(err, service.ErrCreatorNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("PHOTO_NOT_FOUND", "Photo not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to delete photo"))
		return
	}

	c.Status(http.StatusNoContent)
}

// AddVenuePhoto godoc
// @Summary      Добавить фото площадки
// @Description  Привязывает загруженное изображение к галерее площадки
// @Tags         venues
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body map[string]string true "UUID изображения" example({"image_id": "550e8400-e29b-41d4-a716-446655440000"})
// @Success      201 {object} models.VenuePhoto "Фото добавлено"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/venues/photos [post]
func (h *UserHandler) AddVenuePhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	var req struct {
		ImageID string `json:"image_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	photo, err := h.userService.AddVenuePhoto(userID, req.ImageID)
	if err != nil {
		if errors.Is(err, service.ErrVenueNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("VENUE_NOT_FOUND", "Venue profile not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to add photo"))
		return
	}

	c.JSON(http.StatusCreated, photo)
}

// DeleteVenuePhoto godoc
// @Summary      Удалить фото площадки
// @Description  Удаляет фото из галереи площадки (только своё фото)
// @Tags         venues
// @Security     BearerAuth
// @Param        photo_id path int true "ID записи фото"
// @Success      204 "Фото удалено"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/venues/photos/{photo_id} [delete]
func (h *UserHandler) DeleteVenuePhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	photoID, err := strconv.Atoi(c.Param("photo_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid photo ID"))
		return
	}

	if err := h.userService.DeleteVenuePhoto(userID, photoID); err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You can only delete your own photos"))
			return
		}
		if errors.Is(err, service.ErrPhotoNotFound) || errors.Is(err, service.ErrVenueNotFound) {
			c.JSON(http.StatusNotFound, apperror.One("PHOTO_NOT_FOUND", "Photo not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to delete photo"))
		return
	}

	c.Status(http.StatusNoContent)
}

// UploadImage godoc
// @Summary      Загрузить изображение
// @Description  Загружает изображение в MinIO и возвращает метаданные
// @Tags         images
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Файл изображения (jpg, jpeg, png, gif, webp, максимум 10MB)"
// @Param        type formData string true "Тип изображения" Enums(avatar, venue-logo, venue-cover, venue-photo, creator-photo, event-cover)
// @Success      201 {object} models.Image "Изображение загружено"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /users/upload [post]
func (h *UserHandler) UploadImage(c *gin.Context) {
	_, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	imageType := c.PostForm("type")
	if imageType == "" {
		c.JSON(http.StatusBadRequest, apperror.One("FIELD_REQUIRED", "Image type is required (avatar, venue-logo, venue-cover, venue-photo, creator-photo, event-cover)"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("FIELD_REQUIRED", "No file provided"))
		return
	}

	image, err := h.imageService.UploadImage(file, imageType)
	if err != nil {
		if errors.Is(err, service.ErrInvalidFileType) {
			c.JSON(http.StatusBadRequest, apperror.One("INVALID_FILE_TYPE", "Invalid file type, allowed: jpg, jpeg, png, gif, webp"))
			return
		}
		if errors.Is(err, service.ErrFileTooLarge) {
			c.JSON(http.StatusBadRequest, apperror.One("FILE_TOO_LARGE", "File too large, maximum size is 10MB"))
			return
		}
		if errors.Is(err, service.ErrInvalidImageType) {
			c.JSON(http.StatusBadRequest, apperror.One("INVALID_IMAGE_TYPE", "Invalid image type, allowed: avatar, venue-logo, venue-cover, venue-photo, creator-photo, event-cover"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to upload image"))
		return
	}

	c.JSON(http.StatusCreated, image)
}

// GetImage godoc
// @Summary      Получить изображение
// @Description  Возвращает изображение напрямую из хранилища.
// @Tags         images
// @Produce      image/jpeg,image/png,image/gif,image/webp
// @Param        id path string true "UUID изображения"
// @Success      200 {file} binary "Изображение"
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /users/images/{id} [get]
func (h *UserHandler) GetImage(c *gin.Context) {
	imageID := c.Param("id")

	image, data, err := h.imageService.GetImage(imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, apperror.One("IMAGE_NOT_FOUND", "Image not found"))
		return
	}

	c.Header("Content-Type", image.FileType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", image.FileName))
	c.Header("Cache-Control", "public, max-age=31536000")

	c.Data(http.StatusOK, image.FileType, data)
}
