package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type AmenityRepository struct {
	db *gorm.DB
}

func NewAmenityRepository(db *gorm.DB) *AmenityRepository {
	return &AmenityRepository{db: db}
}

func (r *AmenityRepository) Create(amenity *models.Amenity) error {
	return r.db.Create(amenity).Error
}

func (r *AmenityRepository) FindByID(id string) (*models.Amenity, error) {
	var amenity models.Amenity
	err := r.db.Where("id = ?", id).First(&amenity).Error
	if err != nil {
		return nil, err
	}
	return &amenity, nil
}

func (r *AmenityRepository) FindAll(page, perPage int) ([]models.Amenity, int64, error) {
	var amenities []models.Amenity
	var total int64

	r.db.Model(&models.Amenity{}).Count(&total)

	err := r.db.Offset((page - 1) * perPage).Limit(perPage).Find(&amenities).Error
	return amenities, total, err
}

func (r *AmenityRepository) FindByIDs(ids []string) ([]models.Amenity, error) {
	var amenities []models.Amenity
	err := r.db.Where("id IN ?", ids).Find(&amenities).Error
	return amenities, err
}

func (r *AmenityRepository) FindByCategory(category string, page, perPage int) ([]models.Amenity, int64, error) {
	var amenities []models.Amenity
	var total int64

	query := r.db.Where("category = ?", category)
	query.Model(&models.Amenity{}).Count(&total)

	err := query.Offset((page - 1) * perPage).Limit(perPage).Find(&amenities).Error
	return amenities, total, err
}

func (r *AmenityRepository) Update(amenity *models.Amenity) error {
	return r.db.Save(amenity).Error
}

func (r *AmenityRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Amenity{}).Error
}
