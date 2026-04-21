package handlers

import (
	"application-service/internal/apperror"
	"application-service/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CollaborationHandler struct {
	applicationService *service.ApplicationService
}

func NewCollaborationHandler(applicationService *service.ApplicationService) *CollaborationHandler {
	return &CollaborationHandler{applicationService: applicationService}
}

// GetCollaboration получает коллаборацию по ID
// @Summary Get collaboration by ID
// @Tags collaborations
// @Produce json
// @Param id path int true "Collaboration ID"
// @Success 200 {object} models.Collaboration
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /collaborations/{id} [get]
func (h *CollaborationHandler) GetCollaboration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid collaboration ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	collab, err := h.applicationService.GetCollaborationByID(id, userID.(int))
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You are not involved in this collaboration"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("COLLABORATION_NOT_FOUND", "Collaboration not found"))
		return
	}

	c.JSON(http.StatusOK, collab)
}

// ListCollaborations возвращает список коллабораций текущего пользователя
// @Summary List collaborations
// @Tags collaborations
// @Produce json
// @Param status query string false "Collaboration status (pending, completed, cancelled)"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.Collaboration
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /collaborations [get]
func (h *CollaborationHandler) ListCollaborations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	collabs, err := h.applicationService.ListCollaborations(userID.(int), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch collaborations"))
		return
	}

	c.JSON(http.StatusOK, collabs)
}

// ListCollaborationPartners возвращает список уникальных партнёров
// @Summary List collaboration partners
// @Description Get list of unique user IDs this user has successfully collaborated with
// @Tags collaborations
// @Produce json
// @Success 200 {array} int
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /collaborations/partners [get]
func (h *CollaborationHandler) ListCollaborationPartners(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	partnerIDs, err := h.applicationService.ListCollaborationPartners(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch collaboration partners"))
		return
	}

	c.JSON(http.StatusOK, partnerIDs)
}

// GetCompletedEventIDs возвращает ID завершённых мероприятий пользователя
// @Summary Get completed event IDs for a user
// @Description Returns event IDs from completed collaborations for the given user_id. Accessible to any authenticated user.
// @Tags collaborations
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} map[string][]int
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /collaborations/completed-events [get]
func (h *CollaborationHandler) GetCompletedEventIDs(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, apperror.One("MISSING_USER_ID", "user_id query param is required"))
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_USER_ID", "user_id must be a positive integer"))
		return
	}

	eventIDs, err := h.applicationService.GetCompletedEventIDsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch completed event IDs"))
		return
	}

	if eventIDs == nil {
		eventIDs = []int{}
	}

	c.JSON(http.StatusOK, gin.H{"event_ids": eventIDs})
}

// CompleteCollaboration подтверждает проведение мероприятия
// @Summary Complete collaboration (event took place)
// @Description Creator confirms the event took place. Event removed from catalog.
// @Tags collaborations
// @Produce json
// @Param id path int true "Collaboration ID"
// @Success 200 {object} models.Collaboration
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /collaborations/{id}/complete [patch]
func (h *CollaborationHandler) CompleteCollaboration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid collaboration ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "User role not found"))
		return
	}

	collab, err := h.applicationService.CompleteCollaboration(id, userID.(int), role.(string))
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "Only the creator can complete a collaboration"))
			return
		}
		if errors.Is(err, service.ErrCollaborationAlreadyProcessed) {
			c.JSON(http.StatusConflict, apperror.One("COLLABORATION_ALREADY_PROCESSED", "Collaboration has already been completed or cancelled"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("COLLABORATION_NOT_FOUND", "Collaboration not found"))
		return
	}

	c.JSON(http.StatusOK, collab)
}

// CancelCollaboration сообщает, что мероприятие не состоялось
// @Summary Cancel collaboration (event did not take place)
// @Description Creator reports the event did not take place. Event returns to catalog.
// @Tags collaborations
// @Produce json
// @Param id path int true "Collaboration ID"
// @Success 204
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /collaborations/{id}/cancel [patch]
func (h *CollaborationHandler) CancelCollaboration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid collaboration ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "User role not found"))
		return
	}

	if err := h.applicationService.CancelCollaboration(id, userID.(int), role.(string)); err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "Only the creator can cancel a collaboration"))
			return
		}
		if errors.Is(err, service.ErrCollaborationAlreadyProcessed) {
			c.JSON(http.StatusConflict, apperror.One("COLLABORATION_ALREADY_PROCESSED", "Collaboration has already been completed or cancelled"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("COLLABORATION_NOT_FOUND", "Collaboration not found"))
		return
	}

	c.Status(http.StatusNoContent)
}
