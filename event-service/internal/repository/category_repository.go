package repository

import (
	"event-service/internal/models"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) CreateCategory(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) GetCategoryByID(id int) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, id).Error
	return &category, err
}

func (r *CategoryRepository) ListCategories() ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) UpdateCategory(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) DeleteCategory(id int) error {
	return r.db.Delete(&models.Category{}, id).Error
}
