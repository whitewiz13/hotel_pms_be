package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type GuestRepository struct {
	db *gorm.DB
}

func NewGuestRepository(db *gorm.DB) *GuestRepository {
	return &GuestRepository{db: db}
}

func (r *GuestRepository) Create(guest *models.Guest) error {
	return r.db.Create(guest).Error
}

func (r *GuestRepository) CreateInTx(tx *gorm.DB, guest *models.Guest) error {
	return tx.Create(guest).Error
}

func (r *GuestRepository) FindByID(id string) (*models.Guest, error) {
	var guest models.Guest
	if err := r.db.Where("id = ?", id).First(&guest).Error; err != nil {
		return nil, err
	}
	return &guest, nil
}

func (r *GuestRepository) FindByIDAndHotel(id, hotelID string) (*models.Guest, error) {
	var guest models.Guest
	if err := r.db.Where("id = ? AND hotel_id = ?", id, hotelID).First(&guest).Error; err != nil {
		return nil, err
	}
	return &guest, nil
}

// FindByPhone finds a guest by phone number within a hotel.
func (r *GuestRepository) FindByPhone(hotelID, phone string) (*models.Guest, error) {
	var guest models.Guest
	if err := r.db.Where("hotel_id = ? AND phone = ?", hotelID, phone).First(&guest).Error; err != nil {
		return nil, err
	}
	return &guest, nil
}

// FindOrCreate finds an existing guest by phone, or creates a new one.
func (r *GuestRepository) FindOrCreate(tx *gorm.DB, hotelID, name, phone, email string) (*models.Guest, error) {
	// Try to find by phone if provided
	if phone != "" {
		var guest models.Guest
		err := tx.Where("hotel_id = ? AND phone = ?", hotelID, phone).First(&guest).Error
		if err == nil {
			// Update name/email if they've changed
			changed := false
			if name != "" && guest.Name != name {
				guest.Name = name
				changed = true
			}
			if email != "" && guest.Email != email {
				guest.Email = email
				changed = true
			}
			if changed {
				tx.Save(&guest)
			}
			return &guest, nil
		}
	}

	guest := &models.Guest{
		HotelID: hotelID,
		Name:    name,
		Phone:   phone,
		Email:   email,
	}
	if err := tx.Create(guest).Error; err != nil {
		return nil, err
	}
	return guest, nil
}

func (r *GuestRepository) FindByHotelID(hotelID string, page, perPage int) ([]models.Guest, int64, error) {
	var guests []models.Guest
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)

	if err := query.Model(&models.Guest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("name ASC").Offset(offset).Limit(perPage).Find(&guests).Error; err != nil {
		return nil, 0, err
	}

	return guests, total, nil
}

func (r *GuestRepository) Update(guest *models.Guest) error {
	return r.db.Save(guest).Error
}
