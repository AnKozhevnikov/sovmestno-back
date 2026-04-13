package service

import (
	"errors"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewUserService(repo *repository.UserRepository, cfg *config.Config) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
	}
}

// GetMyProfile возвращает профиль текущего пользователя (creator или venue) на основе роли
func (s *UserService) GetMyProfile(userID int) (map[string]interface{}, error) {
	// Получаем базовую информацию о пользователе
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	response := map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	}

	// В зависимости от роли загружаем соответствующий профиль
	switch user.Role {
	case "creator":
		creator, err := s.repo.GetCreatorByUserID(userID)
		if err != nil {
			return nil, ErrCreatorNotFound
		}
		response["profile"] = creator
	case "venue":
		venue, err := s.repo.GetVenueByUserID(userID)
		if err != nil {
			return nil, ErrVenueNotFound
		}
		response["profile"] = venue
	default:
		return nil, ErrUserNotFound
	}

	return response, nil
}

// Creator operations

type CreateCreatorRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=2000"`
	PhotoID     *int   `json:"photo_id"`
	Phone       string `json:"phone" binding:"omitempty,e164"`
	WorkEmail   string `json:"work_email" binding:"omitempty,email"`
	TgPersonal  string `json:"tg_personal_link" binding:"omitempty,url"`
	TgChannel   string `json:"tg_channel_link" binding:"omitempty,url"`
	VkLink      string `json:"vk_link" binding:"omitempty,url"`
	TiktokLink  string `json:"tiktok_link" binding:"omitempty,url"`
	YoutubeLink string `json:"youtube_link" binding:"omitempty,url"`
	DzenLink    string `json:"dzen_link" binding:"omitempty,url"`
}

type UpdateCreatorRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=2000"`
	PhotoID     *int   `json:"photo_id"`
	Phone       string `json:"phone" binding:"omitempty,e164"`
	WorkEmail   string `json:"work_email" binding:"omitempty,email"`
	TgPersonal  string `json:"tg_personal_link" binding:"omitempty,url"`
	TgChannel   string `json:"tg_channel_link" binding:"omitempty,url"`
	VkLink      string `json:"vk_link" binding:"omitempty,url"`
	TiktokLink  string `json:"tiktok_link" binding:"omitempty,url"`
	YoutubeLink string `json:"youtube_link" binding:"omitempty,url"`
	DzenLink    string `json:"dzen_link" binding:"omitempty,url"`
}

func (s *UserService) CreateCreator(userID int, req *CreateCreatorRequest) (*models.Creator, error) {
	// Проверяем, что у пользователя еще нет профиля creator
	existing, _ := s.repo.GetCreatorByUserID(userID)
	if existing != nil {
		return nil, ErrProfileAlreadyExists
	}

	creator := &models.Creator{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		PhotoID:     req.PhotoID,
		Phone:       req.Phone,
		WorkEmail:   req.WorkEmail,
		TgPersonal:  req.TgPersonal,
		TgChannel:   req.TgChannel,
		VkLink:      req.VkLink,
		TiktokLink:  req.TiktokLink,
		YoutubeLink: req.YoutubeLink,
		DzenLink:    req.DzenLink,
	}

	if err := s.repo.CreateCreator(creator); err != nil {
		return nil, err
	}

	return s.repo.GetCreatorByID(creator.ID)
}

func (s *UserService) GetCreator(id int) (*models.Creator, error) {
	return s.repo.GetCreatorByID(id)
}

func (s *UserService) GetCreatorByUserID(userID int) (*models.Creator, error) {
	return s.repo.GetCreatorByUserID(userID)
}

func (s *UserService) UpdateCreator(id, userID int, req *CreateCreatorRequest) (*models.Creator, error) {
	// Проверяем, что creator принадлежит пользователю
	creator, err := s.repo.GetCreatorByID(id)
	if err != nil {
		return nil, err
	}

	if creator.UserID != userID {
		return nil, errors.New("forbidden: not your creator profile")
	}

	// Обновляем поля
	creator.Name = req.Name
	creator.Description = req.Description
	creator.PhotoID = req.PhotoID
	creator.Phone = req.Phone
	creator.WorkEmail = req.WorkEmail
	creator.TgPersonal = req.TgPersonal
	creator.TgChannel = req.TgChannel
	creator.VkLink = req.VkLink
	creator.TiktokLink = req.TiktokLink
	creator.YoutubeLink = req.YoutubeLink
	creator.DzenLink = req.DzenLink

	if err := s.repo.UpdateCreator(creator); err != nil {
		return nil, err
	}

	return s.repo.GetCreatorByID(id)
}

func (s *UserService) UpdateCreatorByUserID(targetUserID, currentUserID int, req *UpdateCreatorRequest) (*models.Creator, error) {
	// Проверяем права доступа
	if targetUserID != currentUserID {
		return nil, ErrAccessDenied
	}

	// Получаем creator по user_id
	creator, err := s.repo.GetCreatorByUserID(targetUserID)
	if err != nil {
		return nil, err
	}

	// Обновляем поля (name только если передан)
	if req.Name != "" {
		creator.Name = req.Name
	}
	creator.Description = req.Description
	creator.PhotoID = req.PhotoID
	creator.Phone = req.Phone
	creator.WorkEmail = req.WorkEmail
	creator.TgPersonal = req.TgPersonal
	creator.TgChannel = req.TgChannel
	creator.VkLink = req.VkLink
	creator.TiktokLink = req.TiktokLink
	creator.YoutubeLink = req.YoutubeLink
	creator.DzenLink = req.DzenLink

	if err := s.repo.UpdateCreator(creator); err != nil {
		return nil, err
	}

	return s.repo.GetCreatorByUserID(targetUserID)
}

func (s *UserService) DeleteCreator(id, userID int) error {
	creator, err := s.repo.GetCreatorByID(id)
	if err != nil {
		return err
	}

	if creator.UserID != userID {
		return ErrAccessDenied
	}

	return s.repo.DeleteCreator(id)
}

func (s *UserService) ListCreators(limit, offset int) ([]models.Creator, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListCreators(limit, offset)
}

func (s *UserService) AddCreatorPhoto(userID, imageID int) (*models.CreatorPhoto, error) {
	creator, err := s.repo.GetCreatorByUserID(userID)
	if err != nil {
		return nil, ErrCreatorNotFound
	}
	return s.repo.AddCreatorPhoto(creator.ID, imageID)
}

func (s *UserService) DeleteCreatorPhoto(userID, photoID int) error {
	creator, err := s.repo.GetCreatorByUserID(userID)
	if err != nil {
		return ErrCreatorNotFound
	}

	photo, err := s.repo.GetCreatorPhoto(photoID)
	if err != nil {
		return ErrPhotoNotFound
	}

	if photo.CreatorID != creator.ID {
		return ErrAccessDenied
	}

	return s.repo.DeleteCreatorPhoto(photoID)
}

func (s *UserService) DeleteCreatorByUserID(targetUserID, currentUserID int) error {
	// Проверяем права доступа
	if targetUserID != currentUserID {
		return ErrAccessDenied
	}

	creator, err := s.repo.GetCreatorByUserID(targetUserID)
	if err != nil {
		return err
	}

	return s.repo.DeleteCreator(creator.ID)
}

// Venue operations

type CreateVenueRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Description   string `json:"description" binding:"omitempty,max=2000"`
	StreetAddress string `json:"street_address" binding:"omitempty,max=500"`
	CityID        *int   `json:"city_id"`
	OpeningHours  string `json:"opening_hours" binding:"omitempty,max=200"`
	Capacity      int    `json:"capacity"`
	LogoID        *int   `json:"logo_id"`
	CoverPhotoID  *int   `json:"cover_photo_id"`
	Phone         string `json:"phone" binding:"omitempty,e164"`
	WorkEmail     string `json:"work_email" binding:"omitempty,email"`
	TgPersonal    string `json:"tg_personal_link" binding:"omitempty,url"`
	TgChannel     string `json:"tg_channel_link" binding:"omitempty,url"`
	VkLink        string `json:"vk_link" binding:"omitempty,url"`
	TiktokLink    string `json:"tiktok_link" binding:"omitempty,url"`
	YoutubeLink   string `json:"youtube_link" binding:"omitempty,url"`
	DzenLink      string `json:"dzen_link" binding:"omitempty,url"`
	CategoryIDs   []int  `json:"category_ids"`
}

type UpdateVenueRequest struct {
	Name          string `json:"name" binding:"omitempty,min=2,max=100"`
	Description   string `json:"description" binding:"omitempty,max=2000"`
	StreetAddress string `json:"street_address" binding:"omitempty,max=500"`
	CityID        *int   `json:"city_id"`
	OpeningHours  string `json:"opening_hours" binding:"omitempty,max=200"`
	Capacity      int    `json:"capacity"`
	LogoID        *int   `json:"logo_id"`
	CoverPhotoID  *int   `json:"cover_photo_id"`
	Phone         string `json:"phone" binding:"omitempty,e164"`
	WorkEmail     string `json:"work_email" binding:"omitempty,email"`
	TgPersonal    string `json:"tg_personal_link" binding:"omitempty,url"`
	TgChannel     string `json:"tg_channel_link" binding:"omitempty,url"`
	VkLink        string `json:"vk_link" binding:"omitempty,url"`
	TiktokLink    string `json:"tiktok_link" binding:"omitempty,url"`
	YoutubeLink   string `json:"youtube_link" binding:"omitempty,url"`
	DzenLink      string `json:"dzen_link" binding:"omitempty,url"`
	CategoryIDs   []int  `json:"category_ids"`
}

func (s *UserService) CreateVenue(userID int, req *CreateVenueRequest) (*models.Venue, error) {
	// Проверяем, что у пользователя еще нет профиля venue
	existing, _ := s.repo.GetVenueByUserID(userID)
	if existing != nil {
		return nil, ErrProfileAlreadyExists
	}

	venue := &models.Venue{
		UserID:        userID,
		Name:          req.Name,
		Description:   req.Description,
		StreetAddress: req.StreetAddress,
		CityID:        req.CityID,
		OpeningHours:  req.OpeningHours,
		Capacity:      req.Capacity,
		LogoID:        req.LogoID,
		CoverPhotoID:  req.CoverPhotoID,
		Phone:         req.Phone,
		WorkEmail:     req.WorkEmail,
		TgPersonal:    req.TgPersonal,
		TgChannel:     req.TgChannel,
		VkLink:        req.VkLink,
		TiktokLink:    req.TiktokLink,
		YoutubeLink:   req.YoutubeLink,
		DzenLink:      req.DzenLink,
	}

	if err := s.repo.CreateVenue(venue); err != nil {
		return nil, err
	}

	// Добавляем категории
	if len(req.CategoryIDs) > 0 {
		if err := s.repo.AddVenueCategories(venue.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Получаем venue с категориями
	result, err := s.repo.GetVenueByID(venue.ID)
	if err != nil {
		return nil, err
	}

	// Получаем категории
	categoryIDs, err := s.repo.GetVenueCategories(venue.ID)
	if err != nil {
		return nil, err
	}
	result.Categories = categoryIDs

	return result, nil
}

func (s *UserService) GetVenue(id int) (*models.Venue, error) {
	venue, err := s.repo.GetVenueByID(id)
	if err != nil {
		return nil, err
	}

	// Получаем категории
	categoryIDs, err := s.repo.GetVenueCategories(id)
	if err != nil {
		return nil, err
	}
	venue.Categories = categoryIDs

	return venue, nil
}

func (s *UserService) GetVenueByUserID(userID int) (*models.Venue, error) {
	venue, err := s.repo.GetVenueByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Получаем категории
	categoryIDs, err := s.repo.GetVenueCategories(venue.ID)
	if err != nil {
		return nil, err
	}
	venue.Categories = categoryIDs

	return venue, nil
}

func (s *UserService) ListVenues(limit, offset int) ([]models.Venue, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // default
	}
	venues, err := s.repo.ListVenues(limit, offset)
	if err != nil {
		return nil, err
	}

	// Получаем категории для каждой площадки
	for i := range venues {
		categoryIDs, err := s.repo.GetVenueCategories(venues[i].ID)
		if err != nil {
			return nil, err
		}
		venues[i].Categories = categoryIDs
	}

	return venues, nil
}

func (s *UserService) UpdateVenue(id, userID int, req *CreateVenueRequest) (*models.Venue, error) {
	venue, err := s.repo.GetVenueByID(id)
	if err != nil {
		return nil, err
	}

	if venue.UserID != userID {
		return nil, errors.New("forbidden: not your venue profile")
	}

	// Обновляем поля
	venue.Name = req.Name
	venue.Description = req.Description
	venue.StreetAddress = req.StreetAddress
	venue.CityID = req.CityID
	venue.OpeningHours = req.OpeningHours
	venue.Capacity = req.Capacity
	venue.LogoID = req.LogoID
	venue.CoverPhotoID = req.CoverPhotoID
	venue.Phone = req.Phone
	venue.WorkEmail = req.WorkEmail
	venue.TgPersonal = req.TgPersonal
	venue.TgChannel = req.TgChannel
	venue.VkLink = req.VkLink
	venue.TiktokLink = req.TiktokLink
	venue.YoutubeLink = req.YoutubeLink
	venue.DzenLink = req.DzenLink

	if err := s.repo.UpdateVenue(venue); err != nil {
		return nil, err
	}

	// Обновляем категории (nil = не трогать, [] = очистить)
	if req.CategoryIDs != nil {
		if err := s.repo.AddVenueCategories(id, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Получаем обновленную venue с категориями
	result, err := s.repo.GetVenueByID(id)
	if err != nil {
		return nil, err
	}

	categoryIDs, err := s.repo.GetVenueCategories(id)
	if err != nil {
		return nil, err
	}
	result.Categories = categoryIDs

	return result, nil
}

func (s *UserService) UpdateVenueByUserID(targetUserID, currentUserID int, req *UpdateVenueRequest) (*models.Venue, error) {
	// Проверяем права доступа
	if targetUserID != currentUserID {
		return nil, ErrAccessDenied
	}

	venue, err := s.repo.GetVenueByUserID(targetUserID)
	if err != nil {
		return nil, err
	}

	// Обновляем поля (name только если передан)
	if req.Name != "" {
		venue.Name = req.Name
	}
	venue.Description = req.Description
	venue.StreetAddress = req.StreetAddress
	venue.CityID = req.CityID
	venue.OpeningHours = req.OpeningHours
	venue.Capacity = req.Capacity
	venue.LogoID = req.LogoID
	venue.CoverPhotoID = req.CoverPhotoID
	venue.Phone = req.Phone
	venue.WorkEmail = req.WorkEmail
	venue.TgPersonal = req.TgPersonal
	venue.TgChannel = req.TgChannel
	venue.VkLink = req.VkLink
	venue.TiktokLink = req.TiktokLink
	venue.YoutubeLink = req.YoutubeLink
	venue.DzenLink = req.DzenLink

	if err := s.repo.UpdateVenue(venue); err != nil {
		return nil, err
	}

	// Обновляем категории (nil = не трогать, [] = очистить)
	if req.CategoryIDs != nil {
		if err := s.repo.AddVenueCategories(venue.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// Получаем обновленную venue с категориями
	result, err := s.repo.GetVenueByUserID(targetUserID)
	if err != nil {
		return nil, err
	}

	categoryIDs, err := s.repo.GetVenueCategories(result.ID)
	if err != nil {
		return nil, err
	}
	result.Categories = categoryIDs

	return result, nil
}

func (s *UserService) DeleteVenue(id, userID int) error {
	venue, err := s.repo.GetVenueByID(id)
	if err != nil {
		return err
	}

	if venue.UserID != userID {
		return ErrAccessDenied
	}

	return s.repo.DeleteVenue(id)
}

func (s *UserService) AddVenuePhoto(userID, imageID int) (*models.VenuePhoto, error) {
	venue, err := s.repo.GetVenueByUserID(userID)
	if err != nil {
		return nil, ErrVenueNotFound
	}
	return s.repo.AddVenuePhoto(venue.ID, imageID)
}

func (s *UserService) DeleteVenuePhoto(userID, photoID int) error {
	venue, err := s.repo.GetVenueByUserID(userID)
	if err != nil {
		return ErrVenueNotFound
	}

	photo, err := s.repo.GetVenuePhoto(photoID)
	if err != nil {
		return ErrPhotoNotFound
	}

	if photo.VenueID != venue.ID {
		return ErrAccessDenied
	}

	return s.repo.DeleteVenuePhoto(photoID)
}

func (s *UserService) DeleteVenueByUserID(targetUserID, currentUserID int) error {
	// Проверяем права доступа
	if targetUserID != currentUserID {
		return ErrAccessDenied
	}

	venue, err := s.repo.GetVenueByUserID(targetUserID)
	if err != nil {
		return err
	}

	return s.repo.DeleteVenue(venue.ID)
}
