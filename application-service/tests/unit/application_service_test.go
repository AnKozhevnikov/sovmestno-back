package unit

import (
	"application-service/internal/service"
	"errors"
	"testing"
)

// ─── CreateApplication ────────────────────────────────────────────────────────

func TestCreateApplication_Success(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	app, err := svc.CreateApplication(
		&service.CreateApplicationRequest{ReceiverID: 2, ReceiverType: "venue", EventID: 10},
		1, "creator",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.SenderID != 1 || app.ReceiverID != 2 || app.EventID != 10 {
		t.Errorf("unexpected app fields: %+v", app)
	}
	if app.Status != "pending" {
		t.Errorf("expected status pending, got %s", app.Status)
	}
}

func TestCreateApplication_CannotApplyToSelf(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	_, err := svc.CreateApplication(
		&service.CreateApplicationRequest{ReceiverID: 1, ReceiverType: "creator", EventID: 10},
		1, "creator",
	)
	if !errors.Is(err, service.ErrCannotApplyToSelf) {
		t.Errorf("expected ErrCannotApplyToSelf, got %v", err)
	}
}

func TestCreateApplication_MirrorExists(t *testing.T) {
	repo := newMockRepo()
	// venue уже отправил заявку creator'у на этот event
	repo.applications[1] = newApp(1, 2, 1, 10, "venue", "creator", "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.CreateApplication(
		&service.CreateApplicationRequest{ReceiverID: 2, ReceiverType: "venue", EventID: 10},
		1, "creator",
	)
	if !errors.Is(err, service.ErrMirrorApplicationExists) {
		t.Errorf("expected ErrMirrorApplicationExists, got %v", err)
	}
}

func TestCreateApplication_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.errCreate = errors.New("db error")
	svc := service.NewApplicationService(repo)

	_, err := svc.CreateApplication(
		&service.CreateApplicationRequest{ReceiverID: 2, ReceiverType: "venue", EventID: 10},
		1, "creator",
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── GetApplicationByID ───────────────────────────────────────────────────────

func TestGetApplicationByID_Success(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	app, err := svc.GetApplicationByID(1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.ID != 1 {
		t.Errorf("expected app ID 1, got %d", app.ID)
	}
}

func TestGetApplicationByID_AccessDenied(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.GetApplicationByID(1, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestGetApplicationByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	_, err := svc.GetApplicationByID(999, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── AcceptApplication ────────────────────────────────────────────────────────

func TestAcceptApplication_Success(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	app, err := svc.AcceptApplication(1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != "accepted" {
		t.Errorf("expected status accepted, got %s", app.Status)
	}
	if len(repo.collaborations) != 1 {
		t.Errorf("expected 1 collaboration, got %d", len(repo.collaborations))
	}
}

func TestAcceptApplication_AccessDenied(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.AcceptApplication(1, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestAcceptApplication_AlreadyProcessed(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "accepted")
	svc := service.NewApplicationService(repo)

	_, err := svc.AcceptApplication(1, 2)
	if !errors.Is(err, service.ErrApplicationAlreadyProcessed) {
		t.Errorf("expected ErrApplicationAlreadyProcessed, got %v", err)
	}
}

func TestAcceptApplication_CollaborationRoles(t *testing.T) {
	// Venue отправил заявку creator'у — venue = sender
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 2, 1, 10, "venue", "creator", "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.AcceptApplication(1, 1) // creator принимает
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	collab := repo.collaborations[1]
	if collab.CreatorUserID != 1 || collab.VenueUserID != 2 {
		t.Errorf("wrong collaboration roles: creatorID=%d venueID=%d", collab.CreatorUserID, collab.VenueUserID)
	}
}

// ─── RejectApplication ────────────────────────────────────────────────────────

func TestRejectApplication_Success(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	app, err := svc.RejectApplication(1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != "rejected" {
		t.Errorf("expected status rejected, got %s", app.Status)
	}
}

func TestRejectApplication_AccessDenied(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.RejectApplication(1, 1) // sender пытается отклонить свою заявку
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestRejectApplication_AlreadyProcessed(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "rejected")
	svc := service.NewApplicationService(repo)

	_, err := svc.RejectApplication(1, 2)
	if !errors.Is(err, service.ErrApplicationAlreadyProcessed) {
		t.Errorf("expected ErrApplicationAlreadyProcessed, got %v", err)
	}
}

// ─── DeleteApplication ────────────────────────────────────────────────────────

func TestDeleteApplication_Success(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	err := svc.DeleteApplication(1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.applications) != 0 {
		t.Error("expected application to be deleted")
	}
}

func TestDeleteApplication_OnlySenderCanDelete(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	svc := service.NewApplicationService(repo)

	err := svc.DeleteApplication(1, 2) // receiver пытается удалить
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestDeleteApplication_OnlyPendingCanBeDeleted(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "accepted")
	svc := service.NewApplicationService(repo)

	err := svc.DeleteApplication(1, 1)
	if !errors.Is(err, service.ErrApplicationAlreadyProcessed) {
		t.Errorf("expected ErrApplicationAlreadyProcessed, got %v", err)
	}
}

// ─── CompleteCollaboration ────────────────────────────────────────────────────

func TestCompleteCollaboration_Success(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	collab, err := svc.CompleteCollaboration(1, 1, "creator")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if collab.Status != "completed" {
		t.Errorf("expected status completed, got %s", collab.Status)
	}
}

func TestCompleteCollaboration_OnlyCreatorRole(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.CompleteCollaboration(1, 2, "venue")
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestCompleteCollaboration_WrongCreator(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.CompleteCollaboration(1, 99, "creator")
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestCompleteCollaboration_AlreadyProcessed(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "completed")
	svc := service.NewApplicationService(repo)

	_, err := svc.CompleteCollaboration(1, 1, "creator")
	if !errors.Is(err, service.ErrCollaborationAlreadyProcessed) {
		t.Errorf("expected ErrCollaborationAlreadyProcessed, got %v", err)
	}
}

// ─── CancelCollaboration ─────────────────────────────────────────────────────

func TestCancelCollaboration_Success(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	err := svc.CancelCollaboration(1, 1, "creator")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCancelCollaboration_OnlyCreatorRole(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	err := svc.CancelCollaboration(1, 2, "venue")
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestCancelCollaboration_AlreadyProcessed(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "cancelled")
	svc := service.NewApplicationService(repo)

	err := svc.CancelCollaboration(1, 1, "creator")
	if !errors.Is(err, service.ErrCollaborationAlreadyProcessed) {
		t.Errorf("expected ErrCollaborationAlreadyProcessed, got %v", err)
	}
}

// ─── GetCollaborationByID ─────────────────────────────────────────────────────

func TestGetCollaborationByID_Success(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	collab, err := svc.GetCollaborationByID(1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if collab.ID != 1 {
		t.Errorf("expected collab ID 1, got %d", collab.ID)
	}
}

func TestGetCollaborationByID_AccessDenied(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	svc := service.NewApplicationService(repo)

	_, err := svc.GetCollaborationByID(1, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

// ─── GetCompletedEventIDs ─────────────────────────────────────────────────────

func TestGetCompletedEventIDsByUserID_AsCreator(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "completed")
	repo.collaborations[2] = newCollab(2, 2, 20, 3, 1, "completed") // user=1 как venue
	repo.collaborations[3] = newCollab(3, 3, 30, 1, 2, "pending")   // не completed
	svc := service.NewApplicationService(repo)

	ids, err := svc.GetCompletedEventIDsByUserID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 event IDs, got %d: %v", len(ids), ids)
	}
}

func TestGetCompletedEventIDsByUserID_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	ids, err := svc.GetCompletedEventIDsByUserID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty result, got %v", ids)
	}
}

// ─── ListCollaborations ───────────────────────────────────────────────────────

func TestListCollaborations_LimitCapped(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	_, err := svc.ListCollaborations(1, "", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.ListCollaborations(1, "", 200, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCollaborations_StatusFilter(t *testing.T) {
	repo := newMockRepo()
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "pending")
	repo.collaborations[2] = newCollab(2, 2, 20, 1, 3, "completed")
	svc := service.NewApplicationService(repo)

	results, err := svc.ListCollaborations(1, "completed", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 completed collaboration, got %d", len(results))
	}
	if results[0].Status != "completed" {
		t.Errorf("expected status completed, got %s", results[0].Status)
	}
}

// ─── ListApplications ────────────────────────────────────────────────────────

func TestListApplications_AsSender(t *testing.T) {
	repo := newMockRepo()
	// userID=1 is sender of 2 apps, receiver of 1
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	repo.applications[2] = newApp(2, 1, 3, 20, "creator", "venue", "accepted")
	repo.applications[3] = newApp(3, 5, 1, 30, "venue", "creator", "pending") // receiver
	svc := service.NewApplicationService(repo)

	results, err := svc.ListApplications(1, "creator", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 applications (sender+receiver), got %d", len(results))
	}
}

func TestListApplications_StatusFilter(t *testing.T) {
	repo := newMockRepo()
	repo.applications[1] = newApp(1, 1, 2, 10, "creator", "venue", "pending")
	repo.applications[2] = newApp(2, 1, 3, 20, "creator", "venue", "accepted")
	svc := service.NewApplicationService(repo)

	results, err := svc.ListApplications(1, "creator", "pending", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 pending application, got %d", len(results))
	}
}

func TestListApplications_LimitCapped(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	_, err := svc.ListApplications(1, "creator", "", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error for limit=0: %v", err)
	}

	_, err = svc.ListApplications(1, "creator", "", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error for limit=500: %v", err)
	}
}

// ─── ListCollaborationPartners ────────────────────────────────────────────────

func TestListCollaborationPartners_Success(t *testing.T) {
	repo := newMockRepo()
	// userID=1 completed collaborations with userID=2 and userID=3
	repo.collaborations[1] = newCollab(1, 1, 10, 1, 2, "completed")
	repo.collaborations[2] = newCollab(2, 2, 20, 3, 1, "completed") // user=1 is venue
	repo.collaborations[3] = newCollab(3, 3, 30, 1, 4, "pending")   // pending — not a partner
	svc := service.NewApplicationService(repo)

	partners, err := svc.ListCollaborationPartners(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(partners) != 2 {
		t.Errorf("expected 2 partners, got %d: %v", len(partners), partners)
	}
}

func TestListCollaborationPartners_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewApplicationService(repo)

	partners, err := svc.ListCollaborationPartners(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(partners) != 0 {
		t.Errorf("expected 0 partners, got %d", len(partners))
	}
}
