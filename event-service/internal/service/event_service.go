package service

import (
	"event-service/internal/models"
	"event-service/internal/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) CreateEvent(req *CreateEventRequest, creatorID int) (*models.Event, error) {
	event := &models.Event{
		CreatorID:    creatorID,
		Title:        req.Title,
		Description:  req.Description,
		CoverPhotoID: req.CoverPhotoID,
	}

	if err := s.repo.CreateEvent(event); err != nil {
		return nil, err
	}

	if len(req.CategoryIDs) > 0 {
		if err := s.repo.AddEventCategories(event.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
		event.Categories = req.CategoryIDs
	}

	return event, nil
}

func (s *EventService) GetEventByID(id int) (*models.Event, error) {
	event, err := s.repo.GetEventByID(id)
	if err != nil {
		return nil, err
	}

	categoryIDs, err := s.repo.GetEventCategories(id)
	if err != nil {
		return nil, err
	}
	event.Categories = categoryIDs

	return event, nil
}

func (s *EventService) ListEvents(creatorID *int, categoryID *int, isActive *bool, isCompleted *bool, limit, offset int) ([]models.Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	events, err := s.repo.ListEvents(creatorID, categoryID, isActive, isCompleted, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range events {
		categoryIDs, err := s.repo.GetEventCategories(events[i].ID)
		if err != nil {
			return nil, err
		}
		events[i].Categories = categoryIDs
	}

	return events, nil
}

func (s *EventService) GetEventsByIDs(ids []int) ([]models.Event, error) {
	if len(ids) == 0 {
		return []models.Event{}, nil
	}
	events, err := s.repo.GetEventsByIDs(ids)
	if err != nil {
		return nil, err
	}

	for i := range events {
		categoryIDs, err := s.repo.GetEventCategories(events[i].ID)
		if err != nil {
			return nil, err
		}
		events[i].Categories = categoryIDs
	}

	return events, nil
}

func (s *EventService) UpdateEvent(id int, req *UpdateEventRequest, creatorID int) (*models.Event, error) {
	event, err := s.repo.GetEventByID(id)
	if err != nil {
		return nil, err
	}

	if event.CreatorID != creatorID {
		return nil, ErrAccessDenied
	}

	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.CoverPhotoID != nil {
		event.CoverPhotoID = req.CoverPhotoID
	}

	if err := s.repo.UpdateEvent(event); err != nil {
		return nil, err
	}

	if req.CategoryIDs != nil {
		if err := s.repo.AddEventCategories(event.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
		event.Categories = req.CategoryIDs
	} else {
		categoryIDs, err := s.repo.GetEventCategories(id)
		if err != nil {
			return nil, err
		}
		event.Categories = categoryIDs
	}

	return event, nil
}

func (s *EventService) PublishEvent(id int, creatorID int) error {
	return s.repo.PublishEvent(id, creatorID)
}

func (s *EventService) DeleteEvent(id int, creatorID int) error {
	event, err := s.repo.GetEventByID(id)
	if err != nil {
		return err
	}

	if event.CreatorID != creatorID {
		return ErrAccessDenied
	}

	return s.repo.DeleteEvent(id)
}


type CreateEventRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	CoverPhotoID *int   `json:"cover_photo_id"`
	CategoryIDs  []int  `json:"category_ids"`
}

type UpdateEventRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	CoverPhotoID *int    `json:"cover_photo_id"`
	CategoryIDs  []int   `json:"category_ids"`
}
