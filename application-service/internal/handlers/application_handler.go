package handlers

import (
	"application-service/internal/apperror"
	"application-service/internal/service"
	"errors"
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
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 409 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /applications [post]
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
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

	var req service.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	app, err := h.applicationService.CreateApplication(&req, userID.(int), role.(string))
	if err != nil {
		if errors.Is(err, service.ErrCannotApplyToSelf) {
			c.JSON(http.StatusBadRequest, apperror.One("CANNOT_APPLY_TO_SELF", "You cannot send an application to yourself"))
			return
		}
		if errors.Is(err, service.ErrDuplicatePendingApplication) {
			c.JSON(http.StatusConflict, apperror.One("DUPLICATE_APPLICATION", "A pending application already exists for this event"))
			return
		}
		if errors.Is(err, service.ErrMirrorApplicationExists) {
			c.JSON(http.StatusConflict, apperror.One("MIRROR_APPLICATION_EXISTS", "An incoming application already exists for this event, check your applications"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create application"))
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
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /applications/{id} [get]
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid application ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	app, err := h.applicationService.GetApplicationByID(id, userID.(int))
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "You are not involved in this application"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("APPLICATION_NOT_FOUND", "Application not found"))
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
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /applications [get]
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	role := c.DefaultQuery("role", "any")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	apps, err := h.applicationService.ListApplications(userID.(int), role, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch applications"))
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
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /applications/{id}/accept [patch]
func (h *ApplicationHandler) AcceptApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid application ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	app, err := h.applicationService.AcceptApplication(id, userID.(int))
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "Only the receiver can accept this application"))
			return
		}
		if errors.Is(err, service.ErrApplicationAlreadyProcessed) {
			c.JSON(http.StatusConflict, apperror.One("APPLICATION_ALREADY_PROCESSED", "Application has already been accepted or rejected"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("APPLICATION_NOT_FOUND", "Application not found"))
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
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /applications/{id}/reject [patch]
func (h *ApplicationHandler) RejectApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid application ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	app, err := h.applicationService.RejectApplication(id, userID.(int))
	if err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "Only the receiver can reject this application"))
			return
		}
		if errors.Is(err, service.ErrApplicationAlreadyProcessed) {
			c.JSON(http.StatusConflict, apperror.One("APPLICATION_ALREADY_PROCESSED", "Application has already been accepted or rejected"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("APPLICATION_NOT_FOUND", "Application not found"))
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
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 401 {object} apperror.ErrorResponse
// @Failure 403 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /applications/{id} [delete]
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid application ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	if err := h.applicationService.DeleteApplication(id, userID.(int)); err != nil {
		if errors.Is(err, service.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, apperror.One("ACCESS_DENIED", "Only the sender can delete this application"))
			return
		}
		if errors.Is(err, service.ErrApplicationAlreadyProcessed) {
			c.JSON(http.StatusConflict, apperror.One("APPLICATION_ALREADY_PROCESSED", "Cannot delete an already processed application"))
			return
		}
		c.JSON(http.StatusNotFound, apperror.One("APPLICATION_NOT_FOUND", "Application not found"))
		return
	}

	c.Status(http.StatusNoContent)
}
