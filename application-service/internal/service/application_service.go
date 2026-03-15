package service

import (
	"application-service/internal/models"
	"application-service/internal/repository"
	"errors"
)

type ApplicationService struct {
	repo *repository.ApplicationRepository
}

func NewApplicationService(repo *repository.ApplicationRepository) *ApplicationService {
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
		return nil, errors.New("cannot send application to yourself")
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
		return nil, errors.New("access denied: you are not involved in this application")
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
		return nil, errors.New("access denied: only receiver can accept application")
	}

	if app.Status != "pending" {
		return nil, errors.New("cannot accept already processed application")
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
		return nil, errors.New("access denied: only receiver can reject application")
	}

	if app.Status != "pending" {
		return nil, errors.New("cannot reject already processed application")
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
		return nil, errors.New("access denied: you are not involved in this collaboration")
	}

	return collab, nil
}

// CompleteCollaboration — creator подтверждает, что мероприятие состоялось.
func (s *ApplicationService) CompleteCollaboration(id int, userID int, userRole string) (*models.Collaboration, error) {
	if userRole != "creator" {
		return nil, errors.New("access denied: only creators can complete collaborations")
	}

	collab, err := s.repo.GetCollaborationByID(id)
	if err != nil {
		return nil, err
	}

	if collab.CreatorUserID != userID {
		return nil, errors.New("access denied: you are not the creator in this collaboration")
	}

	if collab.Status != "pending" {
		return nil, errors.New("can only complete pending collaborations")
	}

	if err := s.repo.CompleteCollaborationTx(collab.ID, collab.EventID); err != nil {
		return nil, err
	}

	collab.Status = "completed"
	return collab, nil
}

// CancelCollaboration — creator сообщает, что мероприятие не состоялось.
func (s *ApplicationService) CancelCollaboration(id int, userID int, userRole string) error {
	if userRole != "creator" {
		return errors.New("access denied: only creators can cancel collaborations")
	}

	collab, err := s.repo.GetCollaborationByID(id)
	if err != nil {
		return err
	}

	if collab.CreatorUserID != userID {
		return errors.New("access denied: you are not the creator in this collaboration")
	}

	if collab.Status != "pending" {
		return errors.New("can only cancel pending collaborations")
	}

	return s.repo.CancelCollaborationTx(collab.ID, collab.EventID)
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

func (s *ApplicationService) DeleteApplication(id int, userID int) error {
	app, err := s.repo.GetApplicationByID(id)
	if err != nil {
		return err
	}

	if app.SenderID != userID {
		return errors.New("access denied: only sender can delete application")
	}

	if app.Status != "pending" {
		return errors.New("cannot delete already processed application")
	}

	return s.repo.DeleteApplication(id)
}
