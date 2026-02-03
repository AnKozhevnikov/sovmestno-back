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

func (r *ApplicationRepository) ListSentApplications(senderID int, status string, limit, offset int) ([]models.Application, error) {
	var applications []models.Application
	query := r.db.Where("sender_id = ?", senderID).
		Order("created_at DESC").
		Limit(limit).Offset(offset)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&applications).Error
	return applications, err
}

func (r *ApplicationRepository) ListReceivedApplications(receiverID int, status string, limit, offset int) ([]models.Application, error) {
	var applications []models.Application
	query := r.db.Where("receiver_id = ?", receiverID).
		Order("created_at DESC").
		Limit(limit).Offset(offset)

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
