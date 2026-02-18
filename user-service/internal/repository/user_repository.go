package repository

import (
	"user-service/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// User operations
func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

// Creator operations
func (r *UserRepository) CreateCreator(creator *models.Creator) error {
	return r.db.Create(creator).Error
}

func (r *UserRepository) GetCreatorByID(id int) (*models.Creator, error) {
	var creator models.Creator
	err := r.db.Preload("Photo").First(&creator, id).Error
	return &creator, err
}

func (r *UserRepository) GetCreatorByUserID(userID int) (*models.Creator, error) {
	var creator models.Creator
	err := r.db.Where("user_id = ?", userID).Preload("Photo").First(&creator).Error
	return &creator, err
}

func (r *UserRepository) UpdateCreator(creator *models.Creator) error {
	return r.db.Save(creator).Error
}

func (r *UserRepository) DeleteCreator(id int) error {
	return r.db.Delete(&models.Creator{}, id).Error
}

// Venue operations
func (r *UserRepository) CreateVenue(venue *models.Venue) error {
	return r.db.Create(venue).Error
}

func (r *UserRepository) GetVenueByID(id int) (*models.Venue, error) {
	var venue models.Venue
	err := r.db.Preload("Logo").
		Preload("CoverPhoto").
		Preload("Photos.Image").
		First(&venue, id).Error
	return &venue, err
}

func (r *UserRepository) GetVenueByUserID(userID int) (*models.Venue, error) {
	var venue models.Venue
	err := r.db.Where("user_id = ?", userID).
		Preload("Logo").
		Preload("CoverPhoto").
		Preload("Photos.Image").
		First(&venue).Error
	return &venue, err
}

func (r *UserRepository) ListVenues(limit, offset int) ([]models.Venue, error) {
	var venues []models.Venue
	err := r.db.Limit(limit).
		Offset(offset).
		Preload("Logo").
		Preload("CoverPhoto").
		Find(&venues).Error
	return venues, err
}

func (r *UserRepository) UpdateVenue(venue *models.Venue) error {
	return r.db.Save(venue).Error
}

func (r *UserRepository) DeleteVenue(id int) error {
	return r.db.Delete(&models.Venue{}, id).Error
}

// Image operations
func (r *UserRepository) CreateImage(image *models.Image) error {
	return r.db.Create(image).Error
}

func (r *UserRepository) GetImageByID(id int) (*models.Image, error) {
	var image models.Image
	err := r.db.First(&image, id).Error
	return &image, err
}

func (r *UserRepository) DeleteImage(id int) error {
	return r.db.Delete(&models.Image{}, id).Error
}

// VenueCategory operations
func (r *UserRepository) AddVenueCategories(venueID int, categoryIDs []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем старые связи
		if err := tx.Where("venue_id = ?", venueID).Delete(&models.VenueCategory{}).Error; err != nil {
			return err
		}
		// Добавляем новые связи
		for _, catID := range categoryIDs {
			if err := tx.Create(&models.VenueCategory{
				VenueID:    venueID,
				CategoryID: catID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) GetVenueCategories(venueID int) ([]int, error) {
	var venueCategories []models.VenueCategory
	err := r.db.Where("venue_id = ?", venueID).Find(&venueCategories).Error
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]int, len(venueCategories))
	for i, vc := range venueCategories {
		categoryIDs[i] = vc.CategoryID
	}
	return categoryIDs, nil
}
