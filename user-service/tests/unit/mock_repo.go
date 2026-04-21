package unit

import (
	"errors"
	"user-service/internal/models"

	"gorm.io/gorm"
)

var errNotFound = errors.New("not found")

type mockUserRepo struct {
	users         map[int]*models.User
	creators      map[int]*models.Creator // keyed by userID
	venues        map[int]*models.Venue   // keyed by userID
	subscriptions map[string]*models.NewsletterSubscription
	favorites     map[int][]int // creatorUserID -> []venueUserID
	nextUserID    int
	nextCreatorID int
	nextVenueID   int
	nextSubID     int

	errCreateUser    error
	errGetByEmail    error
	errCreateCreator error
	errCreateVenue   error
	errAddFav        error
	alreadyFaved     bool
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:         make(map[int]*models.User),
		creators:      make(map[int]*models.Creator),
		venues:        make(map[int]*models.Venue),
		subscriptions: make(map[string]*models.NewsletterSubscription),
		favorites:     make(map[int][]int),
		nextUserID:    1,
		nextCreatorID: 1,
		nextVenueID:   1,
		nextSubID:     1,
	}
}

func (m *mockUserRepo) CreateUser(user *models.User) error {
	if m.errCreateUser != nil {
		return m.errCreateUser
	}
	user.ID = m.nextUserID
	m.nextUserID++
	cp := *user
	m.users[user.ID] = &cp
	return nil
}

func (m *mockUserRepo) GetUserByEmail(email string) (*models.User, error) {
	if m.errGetByEmail != nil {
		return nil, m.errGetByEmail
	}
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) GetUserByID(id int) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) CreateCreator(creator *models.Creator) error {
	if m.errCreateCreator != nil {
		return m.errCreateCreator
	}
	creator.ID = m.nextCreatorID
	m.nextCreatorID++
	cp := *creator
	m.creators[creator.UserID] = &cp
	return nil
}

func (m *mockUserRepo) GetCreatorByID(id int) (*models.Creator, error) {
	for _, c := range m.creators {
		if c.ID == id {
			cp := *c
			return &cp, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) GetCreatorByUserID(userID int) (*models.Creator, error) {
	c, ok := m.creators[userID]
	if !ok {
		return nil, errNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockUserRepo) UpdateCreator(creator *models.Creator) error {
	cp := *creator
	m.creators[creator.UserID] = &cp
	return nil
}

func (m *mockUserRepo) DeleteCreator(id int) error {
	for k, c := range m.creators {
		if c.ID == id {
			delete(m.creators, k)
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) ListCreators(limit, offset int) ([]models.Creator, error) {
	var result []models.Creator
	for _, c := range m.creators {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockUserRepo) AddCreatorPhoto(creatorID int, imageID string) (*models.CreatorPhoto, error) {
	return &models.CreatorPhoto{ID: 1, CreatorID: creatorID, ImageID: imageID}, nil
}

func (m *mockUserRepo) GetCreatorPhoto(photoID int) (*models.CreatorPhoto, error) {
	return nil, errNotFound
}

func (m *mockUserRepo) DeleteCreatorPhoto(photoID int) error { return nil }

func (m *mockUserRepo) CreateVenue(venue *models.Venue) error {
	if m.errCreateVenue != nil {
		return m.errCreateVenue
	}
	venue.ID = m.nextVenueID
	m.nextVenueID++
	cp := *venue
	m.venues[venue.UserID] = &cp
	return nil
}

func (m *mockUserRepo) GetVenueByID(id int) (*models.Venue, error) {
	for _, v := range m.venues {
		if v.ID == id {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) GetVenueByUserID(userID int) (*models.Venue, error) {
	v, ok := m.venues[userID]
	if !ok {
		return nil, errNotFound
	}
	cp := *v
	return &cp, nil
}

func (m *mockUserRepo) ListVenues(limit, offset int) ([]models.Venue, error) {
	var result []models.Venue
	for _, v := range m.venues {
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockUserRepo) UpdateVenue(venue *models.Venue) error {
	cp := *venue
	m.venues[venue.UserID] = &cp
	return nil
}

func (m *mockUserRepo) DeleteVenue(id int) error {
	for k, v := range m.venues {
		if v.ID == id {
			delete(m.venues, k)
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) AddVenuePhoto(venueID int, imageID string) (*models.VenuePhoto, error) {
	return &models.VenuePhoto{ID: 1, VenueID: venueID, ImageID: imageID}, nil
}

func (m *mockUserRepo) GetVenuePhoto(photoID int) (*models.VenuePhoto, error) {
	return nil, errNotFound
}

func (m *mockUserRepo) DeleteVenuePhoto(photoID int) error { return nil }

func (m *mockUserRepo) CreateImage(image *models.Image) error   { return nil }
func (m *mockUserRepo) GetImageByID(id string) (*models.Image, error) {
	return nil, errNotFound
}
func (m *mockUserRepo) DeleteImage(id string) error { return nil }

func (m *mockUserRepo) AddVenueCategories(venueID int, categoryIDs []int) error  { return nil }
func (m *mockUserRepo) GetVenueCategories(venueID int) ([]int, error)             { return nil, nil }

func (m *mockUserRepo) AddCreatorFavoriteVenue(creatorUserID, venueUserID int) (bool, error) {
	if m.errAddFav != nil {
		return false, m.errAddFav
	}
	if m.alreadyFaved {
		return true, nil
	}
	m.favorites[creatorUserID] = append(m.favorites[creatorUserID], venueUserID)
	return false, nil
}

func (m *mockUserRepo) RemoveCreatorFavoriteVenue(creatorUserID, venueUserID int) error {
	list := m.favorites[creatorUserID]
	for i, id := range list {
		if id == venueUserID {
			m.favorites[creatorUserID] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *mockUserRepo) ListCreatorFavoriteVenues(creatorUserID int) ([]models.Venue, error) {
	var result []models.Venue
	for _, venueUserID := range m.favorites[creatorUserID] {
		if v, ok := m.venues[venueUserID]; ok {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (m *mockUserRepo) CreateNewsletterSubscription(sub *models.NewsletterSubscription) error {
	sub.ID = m.nextSubID
	m.nextSubID++
	cp := *sub
	m.subscriptions[sub.Email] = &cp
	return nil
}

func (m *mockUserRepo) GetNewsletterSubscriptionByEmail(email string) (*models.NewsletterSubscription, error) {
	sub, ok := m.subscriptions[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *sub
	return &cp, nil
}

func (m *mockUserRepo) GetNewsletterSubscriptionByToken(token string) (*models.NewsletterSubscription, error) {
	for _, sub := range m.subscriptions {
		if sub.UnsubscribeToken == token {
			cp := *sub
			return &cp, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) DeleteNewsletterSubscription(id int) error {
	for k, sub := range m.subscriptions {
		if sub.ID == id {
			delete(m.subscriptions, k)
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) ListNewsletterSubscriptions() ([]models.NewsletterSubscription, error) {
	var result []models.NewsletterSubscription
	for _, sub := range m.subscriptions {
		result = append(result, *sub)
	}
	return result, nil
}
