package handlers

import (
	"event-service/internal/apperror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListPublicEvents godoc
// @Summary      Публичный каталог мероприятий
// @Description  Возвращает только активные мероприятия (is_active=true). Фильтр is_active игнорируется.
// @Tags         public
// @Produce      json
// @Param        creator_id  query int  false "Фильтр по creator_id"
// @Param        category_id query int  false "Фильтр по категории"
// @Param        limit       query int  false "Количество элементов (по умолчанию 20)"
// @Param        offset      query int  false "Смещение (по умолчанию 0)"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /public/events [get]
func (h *EventHandler) ListPublicEvents(c *gin.Context) {
	var creatorID *int
	if v := c.Query("creator_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err == nil {
			creatorID = &id
		}
	}

	var categoryID *int
	if v := c.Query("category_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err == nil {
			categoryID = &id
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Принудительно только активные мероприятия
	isActive := true

	events, err := h.eventService.ListEvents(creatorID, categoryID, &isActive, nil, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch events"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"limit":  limit,
		"offset": offset,
	})
}

// GetPublicEvent godoc
// @Summary      Публичное мероприятие
// @Description  Возвращает мероприятие по ID. Возвращает 404 если мероприятие неактивно.
// @Tags         public
// @Produce      json
// @Param        id path int true "Event ID"
// @Success      200 {object} models.Event
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /public/events/{id} [get]
func (h *EventHandler) GetPublicEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid event ID"))
		return
	}

	event, err := h.eventService.GetEventByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, apperror.One("EVENT_NOT_FOUND", "Event not found"))
		return
	}

	if !event.IsActive {
		c.JSON(http.StatusNotFound, apperror.One("EVENT_NOT_FOUND", "Event not found"))
		return
	}

	c.JSON(http.StatusOK, event)
}
