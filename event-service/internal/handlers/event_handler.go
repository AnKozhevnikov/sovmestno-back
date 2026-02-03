package handlers

import (
	"event-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// CreateEvent создает новое мероприятие
// @Summary Create event
// @Description Create a new event (creator only)
// @Tags events
// @Accept json
// @Produce json
// @Param event body service.CreateEventRequest true "Event data"
// @Success 201 {object} models.Event
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	// Получаем ID создателя из контекста (устанавливается Gateway)
	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req service.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.CreateEvent(&req, creatorID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GetEvent получает мероприятие по ID
// @Summary Get event by ID
// @Description Get a single event by its ID
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} models.Event
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id} [get]
func (h *EventHandler) GetEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventService.GetEventByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// ListEvents получает список мероприятий
// @Summary List events
// @Description Get list of events with filtering by creator_id, status, and category_id
// @Tags events
// @Produce json
// @Param creator_id query int false "Creator ID"
// @Param status query string false "Event status (published, archived)"
// @Param category_id query int false "Filter by category"
// @Success 200 {array} models.Event
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events [get]
func (h *EventHandler) ListEvents(c *gin.Context) {
	var creatorID *int
	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		id, err := strconv.Atoi(creatorIDStr)
		if err == nil {
			creatorID = &id
		}
	}

	status := c.Query("status")

	var categoryID *int
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	events, err := h.eventService.ListEvents(creatorID, status, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// UpdateEvent обновляет мероприятие
// @Summary Update event
// @Description Update an existing event (creator only)
// @Tags events
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param event body service.UpdateEventRequest true "Event data"
// @Success 200 {object} models.Event
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id} [put]
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req service.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.UpdateEvent(id, &req, creatorID.(int))
	if err != nil {
		if err.Error() == "access denied: you are not the creator of this event" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// DeleteEvent удаляет мероприятие
// @Summary Delete event
// @Description Delete an event by ID (creator only)
// @Tags events
// @Param id path int true "Event ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id} [delete]
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.eventService.DeleteEvent(id, creatorID.(int)); err != nil {
		if err.Error() == "access denied: you are not the creator of this event" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ArchiveEvent переводит мероприятие в статус archived
// @Summary Archive event
// @Description Archive an event (hide from public catalog, creator only)
// @Tags events
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id}/archive [patch]
func (h *EventHandler) ArchiveEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.eventService.ArchiveEvent(id, creatorID.(int)); err != nil {
		if err.Error() == "access denied: you are not the creator of this event" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event archived successfully"})
}

// PublishEvent переводит мероприятие в статус published
// @Summary Publish event
// @Description Publish an event (restore from archive, creator only)
// @Tags events
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id}/publish [patch]
func (h *EventHandler) PublishEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.eventService.PublishEvent(id, creatorID.(int)); err != nil {
		if err.Error() == "access denied: you are not the creator of this event" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event published successfully"})
}
