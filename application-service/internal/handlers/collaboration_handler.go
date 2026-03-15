package handlers

import (
	"application-service/internal/service"
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
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /collaborations/{id} [get]
func (h *CollaborationHandler) GetCollaboration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collaboration ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	collab, err := h.applicationService.GetCollaborationByID(id, userID.(int))
	if err != nil {
		if err.Error() == "access denied: you are not involved in this collaboration" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Collaboration not found"})
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
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /collaborations [get]
func (h *CollaborationHandler) ListCollaborations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	collabs, err := h.applicationService.ListCollaborations(userID.(int), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /collaborations/partners [get]
func (h *CollaborationHandler) ListCollaborationPartners(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	partnerIDs, err := h.applicationService.ListCollaborationPartners(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, partnerIDs)
}

// CompleteCollaboration подтверждает проведение мероприятия
// @Summary Complete collaboration (event took place)
// @Description Creator confirms the event took place. Event removed from catalog.
// @Tags collaborations
// @Produce json
// @Param id path int true "Collaboration ID"
// @Success 200 {object} models.Collaboration
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /collaborations/{id}/complete [patch]
func (h *CollaborationHandler) CompleteCollaboration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collaboration ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	collab, err := h.applicationService.CompleteCollaboration(id, userID.(int), role.(string))
	if err != nil {
		if err.Error() == "access denied: only creators can complete collaborations" ||
			err.Error() == "access denied: you are not the creator in this collaboration" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "can only complete pending collaborations" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Collaboration not found"})
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
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /collaborations/{id}/cancel [patch]
func (h *CollaborationHandler) CancelCollaboration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collaboration ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	if err := h.applicationService.CancelCollaboration(id, userID.(int), role.(string)); err != nil {
		if err.Error() == "access denied: only creators can cancel collaborations" ||
			err.Error() == "access denied: you are not the creator in this collaboration" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "can only cancel pending collaborations" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Collaboration not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
