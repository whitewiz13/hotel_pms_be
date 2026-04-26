package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type RoomTypeRepository struct {
	db *gorm.DB
}

func NewRoomTypeRepository(db *gorm.DB) *RoomTypeRepository {
	return &RoomTypeRepository{db: db}
}

func (r *RoomTypeRepository) Create(roomType *models.RoomType) error {
	return r.db.Create(roomType).Error
}

func (r *RoomTypeRepository) FindByID(id string) (*models.RoomType, error) {
	var roomType models.RoomType
	err := r.db.Where("id = ?", id).First(&roomType).Error
	if err != nil {
		return nil, err
	}
	return &roomType, nil
}

func (r *RoomTypeRepository) FindByHotelID(hotelID string) ([]models.RoomType, error) {
	var roomTypes []models.RoomType
	err := r.db.Where("hotel_id = ?", hotelID).Order("name ASC").Find(&roomTypes).Error
	return roomTypes, err
}

func (r *RoomTypeRepository) FindByIDAndHotel(id, hotelID string) (*models.RoomType, error) {
	var roomType models.RoomType
	err := r.db.Where("id = ? AND hotel_id = ?", id, hotelID).First(&roomType).Error
	if err != nil {
		return nil, err
	}
	return &roomType, nil
}

func (r *RoomTypeRepository) FindByNameAndHotel(name, hotelID string) (*models.RoomType, error) {
	var roomType models.RoomType
	err := r.db.Where("LOWER(name) = LOWER(?) AND hotel_id = ?", name, hotelID).First(&roomType).Error
	if err != nil {
		return nil, err
	}
	return &roomType, nil
}

func (r *RoomTypeRepository) Update(roomType *models.RoomType) error {
	return r.db.Save(roomType).Error
}

func (r *RoomTypeRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.RoomType{}).Error
}
