package handlers

import (
	"application-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	applicationService *service.ApplicationService
}

func NewApplicationHandler(applicationService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{applicationService: applicationService}
}

// CreateApplication создает новую заявку
// @Summary Create application
// @Description Create a new application for collaboration
// @Tags applications
// @Accept json
// @Produce json
// @Param application body service.CreateApplicationRequest true "Application data"
// @Success 201 {object} models.Application
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /applications [post]
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
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

	var req service.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app, err := h.applicationService.CreateApplication(&req, userID.(int), role.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// GetApplication получает заявку по ID
// @Summary Get application by ID
// @Description Get a single application by its ID
// @Tags applications
// @Produce json
// @Param id path int true "Application ID"
// @Success 200 {object} models.Application
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /applications/{id} [get]
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	app, err := h.applicationService.GetApplicationByID(id, userID.(int))
	if err != nil {
		if err.Error() == "access denied: you are not involved in this application" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// ListApplications получает список заявок текущего пользователя
// @Summary List applications
// @Description Get list of applications for current user with optional filters
// @Tags applications
// @Produce json
// @Param role query string false "User's role in application (sender, receiver, any)" default(any)
// @Param status query string false "Application status (pending, accepted, rejected)"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.Application
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /applications [get]
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	role := c.DefaultQuery("role", "any")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	apps, err := h.applicationService.ListApplications(userID.(int), role, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apps)
}

// AcceptApplication принимает заявку
// @Summary Accept application
// @Description Accept a pending application (receiver only). Sets event status to booked.
// @Tags applications
// @Produce json
// @Param id path int true "Application ID"
// @Success 200 {object} models.Application
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /applications/{id}/accept [patch]
func (h *ApplicationHandler) AcceptApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	app, err := h.applicationService.AcceptApplication(id, userID.(int))
	if err != nil {
		if err.Error() == "access denied: only receiver can accept application" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot accept already processed application" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// RejectApplication отклоняет заявку
// @Summary Reject application
// @Description Reject a pending application (receiver only)
// @Tags applications
// @Produce json
// @Param id path int true "Application ID"
// @Success 200 {object} models.Application
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /applications/{id}/reject [patch]
func (h *ApplicationHandler) RejectApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	app, err := h.applicationService.RejectApplication(id, userID.(int))
	if err != nil {
		if err.Error() == "access denied: only receiver can reject application" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot reject already processed application" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// DeleteApplication удаляет заявку
// @Summary Delete application
// @Description Delete an application by ID (sender only, pending status only)
// @Tags applications
// @Param id path int true "Application ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /applications/{id} [delete]
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.applicationService.DeleteApplication(id, userID.(int)); err != nil {
		if err.Error() == "access denied: only sender can delete application" ||
			err.Error() == "cannot delete already processed application" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
