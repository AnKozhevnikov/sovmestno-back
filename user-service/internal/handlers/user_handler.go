package handlers

import (
	"fmt"
	"strconv"
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
// @Failure      400 {object} map[string]string "Ошибка валидации или профиль уже существует"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Router       /users/creators [post]
func (h *UserHandler) CreateCreator(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.CreateCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	creator, err := h.userService.CreateCreator(userID, &req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, creator)
}

// GetCreator godoc
// @Summary      Получить профиль создателя
// @Description  Возвращает профиль создателя по user_id
// @Tags         creators
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      200 {object} models.Creator "Профиль создателя"
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      404 {object} map[string]string "Создатель не найден"
// @Router       /users/creators/{user_id} [get]
func (h *UserHandler) GetCreator(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	creator, err := h.userService.GetCreatorByUserID(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Creator not found"})
		return
	}

	c.JSON(200, creator)
}

// GetMe godoc
// @Summary      Получить мой профиль
// @Description  Возвращает профиль текущего пользователя (создателя или площадки) в зависимости от роли
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Профиль пользователя с ролью и данными профиля"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Профиль не найден"
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	profile, err := h.userService.GetMyProfile(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, profile)
}

// UpdateCreator godoc
// @Summary      Обновить профиль создателя
// @Description  Обновляет профиль создателя (только свой профиль)
// @Tags         creators
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Param        request body service.CreateCreatorRequest true "Обновленные данные профиля"
// @Success      200 {object} models.Creator "Профиль обновлен"
// @Failure      400 {object} map[string]string "Ошибка валидации или запрещено редактировать чужой профиль"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Router       /users/creators/{user_id} [put]
func (h *UserHandler) UpdateCreator(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req service.CreateCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	creator, err := h.userService.UpdateCreatorByUserID(targetUserID, currentUserID, &req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, creator)
}

// DeleteCreator godoc
// @Summary      Удалить профиль создателя
// @Description  Удаляет профиль создателя (только свой профиль)
// @Tags         creators
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      204 "Профиль удален"
// @Failure      400 {object} map[string]string "Некорректный ID или запрещено удалять чужой профиль"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Router       /users/creators/{user_id} [delete]
func (h *UserHandler) DeleteCreator(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.userService.DeleteCreatorByUserID(targetUserID, currentUserID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(204, nil)
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
// @Failure      400 {object} map[string]string "Ошибка валидации или профиль уже существует"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Router       /users/venues [post]
func (h *UserHandler) CreateVenue(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.CreateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	venue, err := h.userService.CreateVenue(userID, &req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, venue)
}

// GetVenue godoc
// @Summary      Получить профиль площадки
// @Description  Возвращает профиль площадки по user_id
// @Tags         venues
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      200 {object} models.Venue "Профиль площадки"
// @Failure      400 {object} map[string]string "Некорректный ID"
// @Failure      404 {object} map[string]string "Площадка не найдена"
// @Router       /users/venues/{user_id} [get]
func (h *UserHandler) GetVenue(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	venue, err := h.userService.GetVenueByUserID(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Venue not found"})
		return
	}

	c.JSON(200, venue)
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
// @Failure      500 {object} map[string]string "Ошибка сервера"
// @Router       /users/venues [get]
func (h *UserHandler) ListVenues(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	venues, err := h.userService.ListVenues(limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch venues"})
		return
	}

	c.JSON(200, gin.H{
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
// @Param        request body service.CreateVenueRequest true "Обновленные данные профиля"
// @Success      200 {object} models.Venue "Профиль обновлен"
// @Failure      400 {object} map[string]string "Ошибка валидации или запрещено редактировать чужой профиль"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Router       /users/venues/{user_id} [put]
func (h *UserHandler) UpdateVenue(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req service.CreateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	venue, err := h.userService.UpdateVenueByUserID(targetUserID, currentUserID, &req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, venue)
}

// DeleteVenue godoc
// @Summary      Удалить профиль площадки
// @Description  Удаляет профиль площадки (только свой профиль)
// @Tags         venues
// @Security     BearerAuth
// @Param        user_id path int true "User ID"
// @Success      204 "Профиль удален"
// @Failure      400 {object} map[string]string "Некорректный ID или запрещено удалять чужой профиль"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Router       /users/venues/{user_id} [delete]
func (h *UserHandler) DeleteVenue(c *gin.Context) {
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.userService.DeleteVenueByUserID(targetUserID, currentUserID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(204, nil)
}

// UploadImage godoc
// @Summary      Загрузить изображение
// @Description  Загружает изображение в MinIO и возвращает метаданные
// @Tags         images
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Файл изображения (jpg, jpeg, png, gif, webp, максимум 10MB)"
// @Param        type formData string true "Тип изображения" Enums(avatar, venue-logo, venue-cover, venue-photo, event-cover)
// @Success      201 {object} models.Image "Изображение загружено"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Ошибка сервера"
// @Router       /users/upload [post]
func (h *UserHandler) UploadImage(c *gin.Context) {
	_, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Получаем тип изображения
	imageType := c.PostForm("type")
	if imageType == "" {
		c.JSON(400, gin.H{"error": "Image type is required (avatar, venue-logo, venue-cover, venue-photo, event-cover)"})
		return
	}

	// Получаем файл из формы
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file provided"})
		return
	}

	// Загружаем в MinIO (бакет определяется автоматически по типу)
	image, err := h.imageService.UploadImage(file, imageType)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, image)
}

// GetImage godoc
// @Summary      Получить изображение
// @Description  Возвращает изображение напрямую из хранилища
// @Tags         images
// @Produce      image/jpeg,image/png,image/gif,image/webp
// @Security     BearerAuth
// @Param        id path int true "ID изображения"
// @Success      200 {file} binary "Изображение"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Изображение не найдено"
// @Router       /users/images/{id} [get]
func (h *UserHandler) GetImage(c *gin.Context) {
	imageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid image ID"})
		return
	}

	// Получаем изображение из MinIO
	image, data, err := h.imageService.GetImage(imageID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем заголовки
	c.Header("Content-Type", image.FileType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", image.FileName))
	c.Header("Cache-Control", "public, max-age=31536000") // Кеширование на 1 год

	// Отправляем данные
	c.Data(200, image.FileType, data)
}
