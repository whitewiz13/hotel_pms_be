package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type HotelRepository struct {
	db *gorm.DB
}

func NewHotelRepository(db *gorm.DB) *HotelRepository {
	return &HotelRepository{db: db}
}

func (r *HotelRepository) Create(hotel *models.Hotel) error {
	return r.db.Create(hotel).Error
}

func (r *HotelRepository) FindByID(id string) (*models.Hotel, error) {
	var hotel models.Hotel
	err := r.db.Where("id = ?", id).First(&hotel).Error
	if err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *HotelRepository) FindAll(page, perPage int) ([]models.Hotel, int64, error) {
	var hotels []models.Hotel
	var total int64

	r.db.Model(&models.Hotel{}).Count(&total)

	err := r.db.Offset((page - 1) * perPage).Limit(perPage).Find(&hotels).Error
	return hotels, total, err
}

func (r *HotelRepository) Update(hotel *models.Hotel) error {
	return r.db.Save(hotel).Error
}

func (r *HotelRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Hotel{}).Error
}

func (r *HotelRepository) FindBySlug(slug string) (*models.Hotel, error) {
	var hotel models.Hotel
	err := r.db.Where("slug = ?", slug).First(&hotel).Error
	if err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *HotelRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Hotel{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}
