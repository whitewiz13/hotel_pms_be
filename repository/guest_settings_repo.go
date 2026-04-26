package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type GuestSettingsRepository struct {
	db *gorm.DB
}

func NewGuestSettingsRepository(db *gorm.DB) *GuestSettingsRepository {
	return &GuestSettingsRepository{db: db}
}

// FindByHotelID returns the guest settings for a hotel, or nil if none exist.
func (r *GuestSettingsRepository) FindByHotelID(hotelID string) (*models.GuestSettings, error) {
	var settings models.GuestSettings
	if err := r.db.Where("hotel_id = ?", hotelID).First(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

// Upsert creates or updates guest settings for a hotel.
func (r *GuestSettingsRepository) Upsert(settings *models.GuestSettings) error {
	var existing models.GuestSettings
	err := r.db.Where("hotel_id = ?", settings.HotelID).First(&existing).Error
	if err == nil {
		// Update existing
		existing.WifiPassword = settings.WifiPassword
		existing.AllowOrders = settings.AllowOrders
		existing.AllowActivities = settings.AllowActivities
		return r.db.Save(&existing).Error
	}
	// Create new
	return r.db.Create(settings).Error
}
