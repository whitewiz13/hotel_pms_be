package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// --- Activity CRUD ---

func (r *ActivityRepository) Create(activity *models.Activity) error {
	return r.db.Create(activity).Error
}

func (r *ActivityRepository) FindByID(id string) (*models.Activity, error) {
	var activity models.Activity
	if err := r.db.Where("id = ?", id).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) FindByIDAndHotel(id, hotelID string) (*models.Activity, error) {
	var activity models.Activity
	if err := r.db.Where("id = ? AND hotel_id = ?", id, hotelID).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) FindByHotelID(hotelID, category string, page, perPage int) ([]models.Activity, int64, error) {
	var activities []models.Activity
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Model(&models.Activity{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("category, name").Offset(offset).Limit(perPage).Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

func (r *ActivityRepository) Update(activity *models.Activity) error {
	return r.db.Save(activity).Error
}

func (r *ActivityRepository) Delete(id string) error {
	return r.db.Delete(&models.Activity{}, "id = ?", id).Error
}

// --- Activity Booking ---

func (r *ActivityRepository) CreateBooking(booking *models.ActivityBooking) error {
	return r.db.Create(booking).Error
}

func (r *ActivityRepository) FindBookingByID(id string) (*models.ActivityBooking, error) {
	var booking models.ActivityBooking
	if err := r.db.Preload("Activity").Preload("Room").Preload("Guest").
		Where("id = ?", id).First(&booking).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *ActivityRepository) FindBookingByIDAndHotel(id, hotelID string) (*models.ActivityBooking, error) {
	var booking models.ActivityBooking
	if err := r.db.Preload("Activity").Preload("Room").Preload("Guest").
		Where("id = ? AND hotel_id = ?", id, hotelID).First(&booking).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *ActivityRepository) FindBookingsByHotelID(hotelID, status, roomID, reservationID, activityID string, page, perPage int) ([]models.ActivityBooking, int64, error) {
	var bookings []models.ActivityBooking
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if roomID != "" {
		query = query.Where("room_id = ?", roomID)
	}
	if reservationID != "" {
		query = query.Where("reservation_id = ?", reservationID)
	}
	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}

	if err := query.Model(&models.ActivityBooking{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Preload("Activity").Preload("Room").Preload("Guest").
		Order("created_at DESC").Offset(offset).Limit(perPage).Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

// FindBookingsByReservationID returns all non-cancelled bookings for billing.
func (r *ActivityRepository) FindBookingsByReservationID(reservationID string) ([]models.ActivityBooking, error) {
	var bookings []models.ActivityBooking
	if err := r.db.Preload("Activity").
		Where("reservation_id = ? AND status != ?", reservationID, models.ActivityBookingCancelled).
		Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *ActivityRepository) UpdateBooking(booking *models.ActivityBooking) error {
	return r.db.Save(booking).Error
}
