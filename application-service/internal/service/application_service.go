package service

import (
	"application-service/internal/models"
	"application-service/internal/repository"
)


type ApplicationService struct {
	repo repository.ApplicationRepositoryInterface
}

func NewApplicationService(repo repository.ApplicationRepositoryInterface) *ApplicationService {
	return &ApplicationService{repo: repo}
}

type CreateApplicationRequest struct {
	ReceiverID   int    `json:"receiver_id" binding:"required"`
	ReceiverType string `json:"receiver_type" binding:"required,oneof=creator venue"`
	EventID      int    `json:"event_id" binding:"required"`
	Message      string `json:"message,omitempty"`
}

func (s *ApplicationService) CreateApplication(req *CreateApplicationRequest, senderID int, senderType string) (*models.Application, error) {
	if senderID == req.ReceiverID && senderType == req.ReceiverType {
		return nil, ErrCannotApplyToSelf
	}

	mirror, err := s.repo.HasMirrorPendingApplication(senderID, req.ReceiverID, req.EventID)
	if err != nil {
		return nil, err
	}
	if mirror {
		return nil, ErrMirrorApplicationExists
	}

	app := &models.Application{
		SenderID:     senderID,
		SenderType:   senderType,
		ReceiverID:   req.ReceiverID,
		ReceiverType: req.ReceiverType,
		EventID:      req.EventID,
		Message:      req.Message,
		Status:       "pending",
	}

	if err := s.repo.CreateApplication(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApplicationService) GetApplicationByID(id int, userID int) (*models.Application, error) {
	app, err := s.repo.GetApplicationByID(id)
	if err != nil {
		return nil, err
	}

	if app.SenderID != userID && app.ReceiverID != userID {
		return nil, ErrAccessDenied
	}

	return app, nil
}

func (s *ApplicationService) ListApplications(userID int, role string, status string, limit, offset int) ([]models.Application, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListApplications(userID, role, status, limit, offset)
}

func (s *ApplicationService) AcceptApplication(id int, userID int) (*models.Application, error) {
	app, err := s.repo.GetApplicationByID(id)
	if err != nil {
		return nil, err
	}

	if app.ReceiverID != userID {
		return nil, ErrAccessDenied
	}

	if app.Status != "pending" {
		return nil, ErrApplicationAlreadyProcessed
	}

	app.Status = "accepted"

	var venueUserID int
	if app.SenderType == "venue" {
		venueUserID = app.SenderID
	} else {
		venueUserID = app.ReceiverID
	}

	var creatorUserID int
	if app.SenderType == "creator" {
		creatorUserID = app.SenderID
	} else {
		creatorUserID = app.ReceiverID
	}

	collab := &models.Collaboration{
		ApplicationID: app.ID,
		EventID:       app.EventID,
		CreatorUserID: creatorUserID,
		VenueUserID:   venueUserID,
		Status:        "pending",
	}

	if err := s.repo.AcceptApplicationTx(app, collab); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApplicationService) RejectApplication(id int, userID int) (*models.Application, error) {
	app, err := s.repo.GetApplicationByID(id)
	if err != nil {
		return nil, err
	}

	if app.ReceiverID != userID {
		return nil, ErrAccessDenied
	}

	if app.Status != "pending" {
		return nil, ErrApplicationAlreadyProcessed
	}

	app.Status = "rejected"
	if err := s.repo.UpdateApplication(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApplicationService) GetCollaborationByID(id int, userID int) (*models.Collaboration, error) {
	collab, err := s.repo.GetCollaborationByID(id)
	if err != nil {
		return nil, err
	}

	if collab.CreatorUserID != userID && collab.VenueUserID != userID {
		return nil, ErrAccessDenied
	}

	return collab, nil
}

func (s *ApplicationService) CompleteCollaboration(id int, userID int, userRole string) (*models.Collaboration, error) {
	if userRole != "creator" {
		return nil, ErrAccessDenied
	}

	collab, err := s.repo.GetCollaborationByID(id)
	if err != nil {
		return nil, err
	}

	if collab.CreatorUserID != userID {
		return nil, ErrAccessDenied
	}

	if collab.Status != "pending" {
		return nil, ErrCollaborationAlreadyProcessed
	}

	if err := s.repo.CompleteCollaborationTx(collab.ID, collab.EventID); err != nil {
		return nil, err
	}

	collab.Status = "completed"
	return collab, nil
}

func (s *ApplicationService) CancelCollaboration(id int, userID int, userRole string) error {
	if userRole != "creator" {
		return ErrAccessDenied
	}

	collab, err := s.repo.GetCollaborationByID(id)
	if err != nil {
		return err
	}

	if collab.CreatorUserID != userID {
		return ErrAccessDenied
	}

	if collab.Status != "pending" {
		return ErrCollaborationAlreadyProcessed
	}

	return s.repo.CancelCollaborationTx(collab.ID)
}

func (s *ApplicationService) ListCollaborationPartners(userID int) ([]int, error) {
	return s.repo.ListCollaborationPartners(userID)
}

func (s *ApplicationService) ListCollaborations(userID int, status string, limit, offset int) ([]models.Collaboration, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListCollaborations(userID, status, limit, offset)
}

func (s *ApplicationService) GetCompletedEventIDsByUserID(userID int) ([]int, error) {
	return s.repo.GetCompletedEventIDsByUserID(userID)
}

func (s *ApplicationService) DeleteApplication(id int, userID int) error {
	app, err := s.repo.GetApplicationByID(id)
	if err != nil {
		return err
	}

	if app.SenderID != userID {
		return ErrAccessDenied
	}

	if app.Status != "pending" {
		return ErrApplicationAlreadyProcessed
	}

	return s.repo.DeleteApplication(id)
}
