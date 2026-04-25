package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room *models.Room) error {
	return r.db.Create(room).Error
}

func (r *RoomRepository) FindByID(id string) (*models.Room, error) {
	var room models.Room
	err := r.db.Preload("Amenities").Where("id = ?", id).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) FindByHotelID(hotelID, status string, page, perPage int) ([]models.Room, int64, error) {
	var rooms []models.Room
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Model(&models.Room{}).Count(&total)

	err := query.Preload("Amenities").Offset((page - 1) * perPage).Limit(perPage).Find(&rooms).Error
	return rooms, total, err
}

func (r *RoomRepository) FindByRoomNumberAndHotel(roomNumber, hotelID string) (*models.Room, error) {
	var room models.Room
	err := r.db.Where("room_number = ? AND hotel_id = ?", roomNumber, hotelID).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) Update(room *models.Room) error {
	return r.db.Save(room).Error
}

func (r *RoomRepository) UpdateAmenities(room *models.Room, amenities []models.Amenity) error {
	return r.db.Model(room).Association("Amenities").Replace(amenities)
}

func (r *RoomRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Room{}).Error
}
