package repository

import "application-service/internal/models"

//go:generate mockgen -source=interface.go -destination=mocks/mock_repository.go -package=mocks

type ApplicationRepositoryInterface interface {
	HasMirrorPendingApplication(senderID, receiverID, eventID int) (bool, error)
	CreateApplication(app *models.Application) error
	GetApplicationByID(id int) (*models.Application, error)
	ListApplications(userID int, role string, status string, limit, offset int) ([]models.Application, error)
	UpdateApplication(app *models.Application) error
	DeleteApplication(id int) error
	AcceptApplicationTx(app *models.Application, collab *models.Collaboration) error
	CompleteCollaborationTx(collaborationID int, eventID int) error
	CancelCollaborationTx(collaborationID int) error
	GetCollaborationByID(id int) (*models.Collaboration, error)
	ListCollaborationPartners(userID int) ([]int, error)
	ListCollaborations(userID int, status string, limit, offset int) ([]models.Collaboration, error)
	GetCompletedEventIDsByUserID(userID int) ([]int, error)
}
