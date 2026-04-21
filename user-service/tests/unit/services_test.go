package unit

import (
	"errors"
	"testing"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// helpers

func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:      "test-secret-key-32-bytes-long!!!",
		AdminSecretKey: "correct-admin-secret",
	}
}

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// ─── AuthService: RegisterCreator ────────────────────────────────────────────

func TestRegisterCreator_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	resp, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email:    "creator@test.com",
		Password: "password123",
		Name:     "Test Creator",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token")
	}
	if resp.User.Role != "creator" {
		t.Errorf("expected role creator, got %s", resp.User.Role)
	}
}

func TestRegisterCreator_EmailAlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	repo.CreateUser(mockUser("creator@test.com", "creator"))
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	_, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email:    "creator@test.com",
		Password: "password123",
		Name:     "Test",
	})
	if !errors.Is(err, service.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

// ─── AuthService: RegisterVenue ──────────────────────────────────────────────

func TestRegisterVenue_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	resp, err := svc.RegisterVenue(&service.RegisterVenueRequest{
		Email:    "venue@test.com",
		Password: "password123",
		Name:     "Test Venue",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.User.Role != "venue" {
		t.Errorf("expected role venue, got %s", resp.User.Role)
	}
}

func TestRegisterVenue_EmailAlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	repo.CreateUser(mockUser("venue@test.com", "venue"))
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	_, err := svc.RegisterVenue(&service.RegisterVenueRequest{
		Email:    "venue@test.com",
		Password: "password123",
		Name:     "Test",
	})
	if !errors.Is(err, service.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

// ─── AuthService: RegisterAdmin ──────────────────────────────────────────────

func TestRegisterAdmin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	resp, err := svc.RegisterAdmin(&service.RegisterAdminRequest{
		Email:       "admin@test.com",
		Password:    "password123",
		AdminSecret: "correct-admin-secret",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.User.Role != "admin" {
		t.Errorf("expected role admin, got %s", resp.User.Role)
	}
}

func TestRegisterAdmin_WrongSecret(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	_, err := svc.RegisterAdmin(&service.RegisterAdminRequest{
		Email:       "admin@test.com",
		Password:    "password123",
		AdminSecret: "wrong-secret",
	})
	if !errors.Is(err, service.ErrInvalidAdminSecret) {
		t.Errorf("expected ErrInvalidAdminSecret, got %v", err)
	}
}

// ─── AuthService: Login ──────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	rdb := newTestRedis(t)
	svc := service.NewAuthService(repo, newTestConfig(), rdb)

	// Сначала регистрируемся чтобы хэш пароля правильный
	_, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email:    "user@test.com",
		Password: "password123",
		Name:     "Test",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resp, err := svc.Login(&service.LoginRequest{
		Email:    "user@test.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	rdb := newTestRedis(t)
	svc := service.NewAuthService(repo, newTestConfig(), rdb)

	svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email:    "user@test.com",
		Password: "password123",
		Name:     "Test",
	})

	_, err := svc.Login(&service.LoginRequest{
		Email:    "user@test.com",
		Password: "wrongpassword",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	_, err := svc.Login(&service.LoginRequest{
		Email:    "nobody@test.com",
		Password: "password123",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// ─── NewsletterService ────────────────────────────────────────────────────────

func TestSubscribe_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewNewsletterService(repo)

	sub, err := svc.Subscribe("user@test.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sub.Email != "user@test.com" {
		t.Errorf("unexpected email: %s", sub.Email)
	}
	if sub.UnsubscribeToken == "" {
		t.Error("expected unsubscribe token to be set")
	}
}

func TestSubscribe_AlreadySubscribed(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewNewsletterService(repo)

	svc.Subscribe("user@test.com")
	_, err := svc.Subscribe("user@test.com")
	if !errors.Is(err, service.ErrAlreadySubscribed) {
		t.Errorf("expected ErrAlreadySubscribed, got %v", err)
	}
}

func TestUnsubscribeByToken_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewNewsletterService(repo)

	sub, _ := svc.Subscribe("user@test.com")
	err := svc.UnsubscribeByToken(sub.UnsubscribeToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.subscriptions) != 0 {
		t.Error("expected subscription to be deleted")
	}
}

func TestUnsubscribeByToken_InvalidToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewNewsletterService(repo)

	err := svc.UnsubscribeByToken("invalid-token")
	if !errors.Is(err, service.ErrInvalidUnsubscribeToken) {
		t.Errorf("expected ErrInvalidUnsubscribeToken, got %v", err)
	}
}

// ─── AuthService: Logout / RefreshAccessToken ────────────────────────────────

func TestLogout_InvalidatesRefreshToken(t *testing.T) {
	repo := newMockUserRepo()
	rdb := newTestRedis(t)
	svc := service.NewAuthService(repo, newTestConfig(), rdb)

	resp, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email:    "user@test.com",
		Password: "password123",
		Name:     "Test",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	svc.Logout(resp.RefreshToken)

	_, err = svc.RefreshAccessToken(resp.RefreshToken)
	if err == nil {
		t.Fatal("expected error after logout, got nil")
	}
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, newTestConfig(), newTestRedis(t))

	_, err := svc.RefreshAccessToken("totally-invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

// ─── UserService: Creator operations ─────────────────────────────────────────

func TestUpdateCreatorByUserID_Success(t *testing.T) {
	repo := newMockUserRepo()
	repo.creators[1] = &models.Creator{ID: 1, UserID: 1, Name: "Old Name"}
	cfg := newTestConfig()
	svc := service.NewUserService(repo, cfg)

	updated, err := svc.UpdateCreatorByUserID(1, 1, &service.UpdateCreatorRequest{Name: "New Name"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %s", updated.Name)
	}
}

func TestUpdateCreatorByUserID_AccessDenied(t *testing.T) {
	repo := newMockUserRepo()
	repo.creators[1] = &models.Creator{ID: 1, UserID: 1, Name: "Test"}
	svc := service.NewUserService(repo, newTestConfig())

	_, err := svc.UpdateCreatorByUserID(1, 99, &service.UpdateCreatorRequest{Name: "Hacked"})
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestDeleteCreatorByUserID_Success(t *testing.T) {
	repo := newMockUserRepo()
	repo.creators[1] = &models.Creator{ID: 1, UserID: 1, Name: "Test"}
	svc := service.NewUserService(repo, newTestConfig())

	err := svc.DeleteCreatorByUserID(1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.creators) != 0 {
		t.Error("expected creator to be deleted")
	}
}

func TestDeleteCreatorByUserID_AccessDenied(t *testing.T) {
	repo := newMockUserRepo()
	repo.creators[1] = &models.Creator{ID: 1, UserID: 1, Name: "Test"}
	svc := service.NewUserService(repo, newTestConfig())

	err := svc.DeleteCreatorByUserID(1, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

// ─── UserService: Venue operations ───────────────────────────────────────────

func TestUpdateVenueByUserID_Success(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[1] = &models.Venue{ID: 1, UserID: 1, Name: "Old Name"}
	svc := service.NewUserService(repo, newTestConfig())

	updated, err := svc.UpdateVenueByUserID(1, 1, &service.UpdateVenueRequest{Name: "New Name"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %s", updated.Name)
	}
}

func TestUpdateVenueByUserID_AccessDenied(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[1] = &models.Venue{ID: 1, UserID: 1, Name: "Test"}
	svc := service.NewUserService(repo, newTestConfig())

	_, err := svc.UpdateVenueByUserID(1, 99, &service.UpdateVenueRequest{Name: "Hacked"})
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestDeleteVenueByUserID_Success(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[1] = &models.Venue{ID: 1, UserID: 1, Name: "Test"}
	svc := service.NewUserService(repo, newTestConfig())

	err := svc.DeleteVenueByUserID(1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.venues) != 0 {
		t.Error("expected venue to be deleted")
	}
}

func TestDeleteVenueByUserID_AccessDenied(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[1] = &models.Venue{ID: 1, UserID: 1, Name: "Test"}
	svc := service.NewUserService(repo, newTestConfig())

	err := svc.DeleteVenueByUserID(1, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

// ─── UserService: ProfileAlreadyExists ───────────────────────────────────────

func TestCreateCreator_ProfileAlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	repo.creators[1] = &models.Creator{ID: 1, UserID: 1, Name: "Existing"}
	svc := service.NewUserService(repo, newTestConfig())

	_, err := svc.CreateCreator(1, &service.CreateCreatorRequest{Name: "New"})
	if !errors.Is(err, service.ErrProfileAlreadyExists) {
		t.Errorf("expected ErrProfileAlreadyExists, got %v", err)
	}
}

func TestCreateVenue_ProfileAlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[1] = &models.Venue{ID: 1, UserID: 1, Name: "Existing"}
	svc := service.NewUserService(repo, newTestConfig())

	_, err := svc.CreateVenue(1, &service.CreateVenueRequest{Name: "New"})
	if !errors.Is(err, service.ErrProfileAlreadyExists) {
		t.Errorf("expected ErrProfileAlreadyExists, got %v", err)
	}
}

// ─── FavoritesService ────────────────────────────────────────────────────────

func TestAddFavoriteVenue_Success(t *testing.T) {
	repo := newMockUserRepo()
	// Venue с userID=2 должна существовать
	repo.venues[2] = newVenue(2, 2, "Test Venue")
	svc := service.NewFavoritesService(repo)

	err := svc.AddFavoriteVenue(1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.favorites[1]) != 1 {
		t.Error("expected venue to be added to favorites")
	}
}

func TestAddFavoriteVenue_VenueNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewFavoritesService(repo)

	err := svc.AddFavoriteVenue(1, 999)
	if !errors.Is(err, service.ErrVenueNotFound) {
		t.Errorf("expected ErrVenueNotFound, got %v", err)
	}
}

func TestAddFavoriteVenue_AlreadyFavorited(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[2] = newVenue(2, 2, "Test Venue")
	repo.alreadyFaved = true
	svc := service.NewFavoritesService(repo)

	err := svc.AddFavoriteVenue(1, 2)
	if !errors.Is(err, service.ErrAlreadyFavorited) {
		t.Errorf("expected ErrAlreadyFavorited, got %v", err)
	}
}

func TestRemoveFavoriteVenue_Success(t *testing.T) {
	repo := newMockUserRepo()
	repo.venues[2] = newVenue(2, 2, "Test Venue")
	repo.favorites[1] = []int{2}
	svc := service.NewFavoritesService(repo)

	err := svc.RemoveFavoriteVenue(1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.favorites[1]) != 0 {
		t.Error("expected favorite to be removed")
	}
}

func TestRemoveFavoriteVenue_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewFavoritesService(repo)

	err := svc.RemoveFavoriteVenue(1, 999)
	if !errors.Is(err, service.ErrFavoriteNotFound) {
		t.Errorf("expected ErrFavoriteNotFound, got %v", err)
	}
}
