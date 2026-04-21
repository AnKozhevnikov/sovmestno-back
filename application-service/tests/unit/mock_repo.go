package unit

import (
	"application-service/internal/models"
	"errors"
)

var errNotFound = errors.New("not found")

// mockRepo — ручной мок репозитория для unit-тестов
type mockRepo struct {
	applications   map[int]*models.Application
	nextAppID      int
	collaborations map[int]*models.Collaboration
	nextCollabID   int

	errCreate      error
	errGetApp      error
	errMirror      bool
	errAcceptTx    error
	errCompleteTx  error
	errCancelTx    error
	errGetCollab   error
	errCompletedIDs error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		applications:   make(map[int]*models.Application),
		collaborations: make(map[int]*models.Collaboration),
		nextAppID:      1,
		nextCollabID:   1,
	}
}

func (m *mockRepo) HasMirrorPendingApplication(senderID, receiverID, eventID int) (bool, error) {
	if m.errMirror {
		return true, nil
	}
	for _, app := range m.applications {
		if app.SenderID == receiverID && app.ReceiverID == senderID &&
			app.EventID == eventID && app.Status == "pending" {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) CreateApplication(app *models.Application) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	app.ID = m.nextAppID
	m.nextAppID++
	m.applications[app.ID] = app
	return nil
}

func (m *mockRepo) GetApplicationByID(id int) (*models.Application, error) {
	if m.errGetApp != nil {
		return nil, m.errGetApp
	}
	app, ok := m.applications[id]
	if !ok {
		return nil, errNotFound
	}
	copy := *app
	return &copy, nil
}

func (m *mockRepo) ListApplications(userID int, role string, status string, limit, offset int) ([]models.Application, error) {
	var result []models.Application
	for _, app := range m.applications {
		if app.SenderID == userID || app.ReceiverID == userID {
			if status == "" || app.Status == status {
				result = append(result, *app)
			}
		}
	}
	return result, nil
}

func (m *mockRepo) UpdateApplication(app *models.Application) error {
	m.applications[app.ID] = app
	return nil
}

func (m *mockRepo) DeleteApplication(id int) error {
	delete(m.applications, id)
	return nil
}

func (m *mockRepo) AcceptApplicationTx(app *models.Application, collab *models.Collaboration) error {
	if m.errAcceptTx != nil {
		return m.errAcceptTx
	}
	m.applications[app.ID] = app
	collab.ID = m.nextCollabID
	m.nextCollabID++
	m.collaborations[collab.ID] = collab
	return nil
}

func (m *mockRepo) CompleteCollaborationTx(collaborationID int, eventID int) error {
	if m.errCompleteTx != nil {
		return m.errCompleteTx
	}
	if c, ok := m.collaborations[collaborationID]; ok {
		c.Status = "completed"
	}
	return nil
}

func (m *mockRepo) CancelCollaborationTx(collaborationID int) error {
	if m.errCancelTx != nil {
		return m.errCancelTx
	}
	if c, ok := m.collaborations[collaborationID]; ok {
		c.Status = "cancelled"
	}
	return nil
}

func (m *mockRepo) GetCollaborationByID(id int) (*models.Collaboration, error) {
	if m.errGetCollab != nil {
		return nil, m.errGetCollab
	}
	c, ok := m.collaborations[id]
	if !ok {
		return nil, errNotFound
	}
	copy := *c
	return &copy, nil
}

func (m *mockRepo) ListCollaborationPartners(userID int) ([]int, error) {
	seen := make(map[int]bool)
	var result []int
	for _, c := range m.collaborations {
		if c.Status != "completed" {
			continue
		}
		if c.CreatorUserID == userID && !seen[c.VenueUserID] {
			seen[c.VenueUserID] = true
			result = append(result, c.VenueUserID)
		} else if c.VenueUserID == userID && !seen[c.CreatorUserID] {
			seen[c.CreatorUserID] = true
			result = append(result, c.CreatorUserID)
		}
	}
	return result, nil
}

func (m *mockRepo) ListCollaborations(userID int, status string, limit, offset int) ([]models.Collaboration, error) {
	var result []models.Collaboration
	for _, c := range m.collaborations {
		if c.CreatorUserID == userID || c.VenueUserID == userID {
			if status == "" || c.Status == status {
				result = append(result, *c)
			}
		}
	}
	return result, nil
}

func (m *mockRepo) GetCompletedEventIDsByUserID(userID int) ([]int, error) {
	if m.errCompletedIDs != nil {
		return nil, m.errCompletedIDs
	}
	var result []int
	for _, c := range m.collaborations {
		if c.Status == "completed" && (c.CreatorUserID == userID || c.VenueUserID == userID) {
			result = append(result, c.EventID)
		}
	}
	return result, nil
}

// helpers

func newApp(id, senderID, receiverID, eventID int, senderType, receiverType, status string) *models.Application {
	return &models.Application{
		ID:           id,
		SenderID:     senderID,
		SenderType:   senderType,
		ReceiverID:   receiverID,
		ReceiverType: receiverType,
		EventID:      eventID,
		Status:       status,
	}
}

func newCollab(id, appID, eventID, creatorID, venueID int, status string) *models.Collaboration {
	return &models.Collaboration{
		ID:            id,
		ApplicationID: appID,
		EventID:       eventID,
		CreatorUserID: creatorID,
		VenueUserID:   venueID,
		Status:        status,
	}
}
