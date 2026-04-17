package handlers

import (
	"errors"
	"net/http"
	"user-service/internal/apperror"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type NewsletterHandler struct {
	newsletterService *service.NewsletterService
}

func NewNewsletterHandler(newsletterService *service.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{newsletterService: newsletterService}
}

type subscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Subscribe godoc
// @Summary      Subscribe to newsletter
// @Description  Subscribe an email address to the newsletter
// @Tags         newsletter
// @Accept       json
// @Produce      json
// @Param        request body subscribeRequest true "Email to subscribe"
// @Success      201 {object} models.NewsletterSubscription
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /newsletter/subscribe [post]
func (h *NewsletterHandler) Subscribe(c *gin.Context) {
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	sub, err := h.newsletterService.Subscribe(req.Email)
	if err != nil {
		if errors.Is(err, service.ErrAlreadySubscribed) {
			c.JSON(http.StatusConflict, apperror.One("ALREADY_SUBSCRIBED", "This email is already subscribed"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to subscribe"))
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// Unsubscribe godoc
// @Summary      Unsubscribe from newsletter
// @Description  Unsubscribe using the token from the email link
// @Tags         newsletter
// @Produce      json
// @Param        token query string true "Unsubscribe token"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Router       /newsletter/unsubscribe [get]
func (h *NewsletterHandler) Unsubscribe(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, apperror.One("FIELD_REQUIRED", "Field 'token' is required"))
		return
	}

	if err := h.newsletterService.UnsubscribeByToken(token); err != nil {
		if errors.Is(err, service.ErrInvalidUnsubscribeToken) {
			c.JSON(http.StatusBadRequest, apperror.One("INVALID_UNSUBSCRIBE_TOKEN", "Invalid or expired unsubscribe token"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to unsubscribe"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unsubscribed"})
}

// ListSubscriptions godoc
// @Summary      List newsletter subscribers
// @Description  Get all newsletter subscribers (admin only)
// @Tags         newsletter
// @Produce      json
// @Success      200 {array} models.NewsletterSubscription
// @Failure      403 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /newsletter/subscribers [get]
func (h *NewsletterHandler) ListSubscriptions(c *gin.Context) {
	subs, err := h.newsletterService.ListSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch subscribers"))
		return
	}

	c.JSON(http.StatusOK, subs)
}
