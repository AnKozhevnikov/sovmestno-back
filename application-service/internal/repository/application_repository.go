package repository

import (
	"application-service/internal/models"
	"gorm.io/gorm"
)

type ApplicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) CreateApplication(app *models.Application) error {
	return r.db.Create(app).Error
}

func (r *ApplicationRepository) GetApplicationByID(id int) (*models.Application, error) {
	var app models.Application
	err := r.db.First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *ApplicationRepository) ListApplications(userID int, role string, status string, limit, offset int) ([]models.Application, error) {
	var applications []models.Application

	query := r.db.Order("updated_at DESC").Limit(limit).Offset(offset)

	switch role {
	case "sender":
		query = query.Where("sender_id = ?", userID)
	case "receiver":
		query = query.Where("receiver_id = ?", userID)
	default: // any
		query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&applications).Error
	return applications, err
}

func (r *ApplicationRepository) UpdateApplication(app *models.Application) error {
	return r.db.Save(app).Error
}

func (r *ApplicationRepository) DeleteApplication(id int) error {
	return r.db.Delete(&models.Application{}, id).Error
}

func (r *ApplicationRepository) ListCollaborations(userID int, limit, offset int) ([]models.Application, error) {
	var applications []models.Application
	err := r.db.Where("(sender_id = ? OR receiver_id = ?) AND status = 'published'", userID, userID).
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&applications).Error
	return applications, err
}
