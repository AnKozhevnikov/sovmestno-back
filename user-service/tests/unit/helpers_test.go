package unit

import "user-service/internal/models"

func mockUser(email, role string) *models.User {
	return &models.User{Email: email, PasswordHash: "hash", Role: role}
}

func newVenue(id, userID int, name string) *models.Venue {
	return &models.Venue{ID: id, UserID: userID, Name: name}
}
