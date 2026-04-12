package handlers

import (
	"event-service/internal/apperror"
	"event-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// CreateCategory создает новую категорию
// @Summary Create category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body service.CreateCategoryRequest true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create category"))
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetCategory получает категорию по ID
// @Summary Get category by ID
// @Description Get a single category by its ID
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid category ID"))
		return
	}

	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, apperror.One("CATEGORY_NOT_FOUND", "Category not found"))
		return
	}

	c.JSON(http.StatusOK, category)
}

// ListCategories получает список категорий
// @Summary List categories
// @Description Get list of all categories
// @Tags categories
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to fetch categories"))
		return
	}

	c.JSON(http.StatusOK, categories)
}

// UpdateCategory обновляет категорию
// @Summary Update category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body service.UpdateCategoryRequest true "Category data"
// @Success 200 {object} models.Category
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 404 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid category ID"))
		return
	}

	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	category, err := h.categoryService.UpdateCategory(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to update category"))
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory удаляет категорию
// @Summary Delete category
// @Description Delete a category by ID
// @Tags categories
// @Param id path int true "Category ID"
// @Success 204
// @Failure 400 {object} apperror.ErrorResponse
// @Failure 500 {object} apperror.ErrorResponse
// @Security BearerAuth
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_ID", "Invalid category ID"))
		return
	}

	if err := h.categoryService.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to delete category"))
		return
	}

	c.Status(http.StatusNoContent)
}
