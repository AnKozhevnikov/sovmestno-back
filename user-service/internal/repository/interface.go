package repository

import "user-service/internal/models"

type UserRepositoryInterface interface {
	// User
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)

	// Creator
	CreateCreator(creator *models.Creator) error
	GetCreatorByID(id int) (*models.Creator, error)
	GetCreatorByUserID(userID int) (*models.Creator, error)
	UpdateCreator(creator *models.Creator) error
	DeleteCreator(id int) error
	ListCreators(limit, offset int) ([]models.Creator, error)

	// CreatorPhoto
	AddCreatorPhoto(creatorID int, imageID string) (*models.CreatorPhoto, error)
	GetCreatorPhoto(photoID int) (*models.CreatorPhoto, error)
	DeleteCreatorPhoto(photoID int) error

	// Venue
	CreateVenue(venue *models.Venue) error
	GetVenueByID(id int) (*models.Venue, error)
	GetVenueByUserID(userID int) (*models.Venue, error)
	ListVenues(limit, offset int) ([]models.Venue, error)
	UpdateVenue(venue *models.Venue) error
	DeleteVenue(id int) error

	// VenuePhoto
	AddVenuePhoto(venueID int, imageID string) (*models.VenuePhoto, error)
	GetVenuePhoto(photoID int) (*models.VenuePhoto, error)
	DeleteVenuePhoto(photoID int) error

	// Image
	CreateImage(image *models.Image) error
	GetImageByID(id string) (*models.Image, error)
	DeleteImage(id string) error

	// VenueCategory
	AddVenueCategories(venueID int, categoryIDs []int) error
	GetVenueCategories(venueID int) ([]int, error)

	// Favorites
	AddCreatorFavoriteVenue(creatorUserID, venueUserID int) (bool, error)
	RemoveCreatorFavoriteVenue(creatorUserID, venueUserID int) error
	ListCreatorFavoriteVenues(creatorUserID int) ([]models.Venue, error)

	// Newsletter
	CreateNewsletterSubscription(sub *models.NewsletterSubscription) error
	GetNewsletterSubscriptionByEmail(email string) (*models.NewsletterSubscription, error)
	GetNewsletterSubscriptionByToken(token string) (*models.NewsletterSubscription, error)
	DeleteNewsletterSubscription(id int) error
	ListNewsletterSubscriptions() ([]models.NewsletterSubscription, error)
}
