package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) Create(item *models.MenuItem) error {
	return r.db.Create(item).Error
}

func (r *MenuRepository) FindByID(id string) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MenuRepository) FindByIDAndHotel(id, hotelID string) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := r.db.Where("id = ? AND hotel_id = ?", id, hotelID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MenuRepository) FindByHotelID(hotelID, category string, page, perPage int) ([]models.MenuItem, int64, error) {
	var items []models.MenuItem
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Model(&models.MenuItem{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("category, name").Offset(offset).Limit(perPage).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *MenuRepository) Update(item *models.MenuItem) error {
	return r.db.Save(item).Error
}

func (r *MenuRepository) Delete(id string) error {
	return r.db.Delete(&models.MenuItem{}, "id = ?", id).Error
}
