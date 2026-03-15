package repository

import (
	"application-service/internal/models"
	"time"

	"gorm.io/gorm"
)

// Event — локальная структура для обновления статуса события напрямую в общей БД
type Event struct {
	ID        int       `gorm:"primaryKey"`
	Status    string    `gorm:"column:status"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Event) TableName() string { return "events" }

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

// AcceptApplicationTx атомарно принимает заявку и создаёт коллаборацию.
// Если событие active — переводит в booked. Если уже booked/completed — оставляет как есть.
func (r *ApplicationRepository) AcceptApplicationTx(app *models.Application, collab *models.Collaboration) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var currentEventStatus string
		if err := tx.Model(&Event{}).Select("status").Where("id = ?", app.EventID).Scan(&currentEventStatus).Error; err != nil {
			return err
		}

		if err := tx.Save(app).Error; err != nil {
			return err
		}

		if currentEventStatus == "active" {
			if err := tx.Model(&Event{}).Where("id = ?", app.EventID).Update("status", "booked").Error; err != nil {
				return err
			}
		}

		return tx.Create(collab).Error
	})
}

// CompleteCollaborationTx атомарно переводит коллаборацию и событие в completed.
func (r *ApplicationRepository) CompleteCollaborationTx(collaborationID int, eventID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Collaboration{}).Where("id = ?", collaborationID).Update("status", "completed").Error; err != nil {
			return err
		}
		return tx.Model(&Event{}).Where("id = ?", eventID).Update("status", "completed").Error
	})
}

// CancelCollaborationTx атомарно отменяет коллаборацию.
// Если событие booked и больше нет других pending коллабораций — возвращает событие в active.
func (r *ApplicationRepository) CancelCollaborationTx(collaborationID int, eventID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Collaboration{}).Where("id = ?", collaborationID).Update("status", "cancelled").Error; err != nil {
			return err
		}

		var currentEventStatus string
		if err := tx.Model(&Event{}).Select("status").Where("id = ?", eventID).Scan(&currentEventStatus).Error; err != nil {
			return err
		}

		if currentEventStatus == "booked" {
			var pendingCount int64
			if err := tx.Model(&models.Collaboration{}).
				Where("event_id = ? AND status = 'pending'", eventID).
				Count(&pendingCount).Error; err != nil {
				return err
			}

			if pendingCount == 0 {
				return tx.Model(&Event{}).Where("id = ?", eventID).Update("status", "active").Error
			}
		}

		return nil
	})
}

func (r *ApplicationRepository) GetCollaborationByID(id int) (*models.Collaboration, error) {
	var collab models.Collaboration
	err := r.db.First(&collab, id).Error
	if err != nil {
		return nil, err
	}
	return &collab, nil
}

func (r *ApplicationRepository) ListCollaborationPartners(userID int) ([]int, error) {
	var partnerIDs []int
	err := r.db.Raw(`
		SELECT DISTINCT
			CASE WHEN creator_user_id = ? THEN venue_user_id ELSE creator_user_id END AS partner_id
		FROM collaborations
		WHERE (creator_user_id = ? OR venue_user_id = ?) AND status = 'completed'
	`, userID, userID, userID).Scan(&partnerIDs).Error
	return partnerIDs, err
}

func (r *ApplicationRepository) ListCollaborations(userID int, status string, limit, offset int) ([]models.Collaboration, error) {
	var collabs []models.Collaboration
	query := r.db.Where("creator_user_id = ? OR venue_user_id = ?", userID, userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&collabs).Error
	return collabs, err
}
