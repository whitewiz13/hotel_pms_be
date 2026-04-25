package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type BillRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) *BillRepository {
	return &BillRepository{db: db}
}

func (r *BillRepository) Create(bill *models.Bill) error {
	return r.db.Create(bill).Error
}

func (r *BillRepository) CreateLineItems(items []models.BillLineItem) error {
	return r.db.Create(&items).Error
}

func (r *BillRepository) FindByID(id string) (*models.Bill, error) {
	var bill models.Bill
	if err := r.db.Preload("LineItems").Preload("Room").Preload("Reservation").
		Where("id = ?", id).First(&bill).Error; err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *BillRepository) FindByIDAndHotel(id, hotelID string) (*models.Bill, error) {
	var bill models.Bill
	if err := r.db.Preload("LineItems").Preload("Room").Preload("Reservation").
		Where("id = ? AND hotel_id = ?", id, hotelID).First(&bill).Error; err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *BillRepository) FindByReservationID(reservationID string) (*models.Bill, error) {
	var bill models.Bill
	if err := r.db.Preload("LineItems").Preload("Room").Preload("Reservation").
		Where("reservation_id = ?", reservationID).First(&bill).Error; err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *BillRepository) FindByHotelID(hotelID, status, reservationID string, page, perPage int) ([]models.Bill, int64, error) {
	var bills []models.Bill
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if reservationID != "" {
		query = query.Where("reservation_id = ?", reservationID)
	}

	if err := query.Model(&models.Bill{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Preload("LineItems").Preload("Room").Preload("Reservation").
		Order("created_at DESC").Offset(offset).Limit(perPage).Find(&bills).Error; err != nil {
		return nil, 0, err
	}

	return bills, total, nil
}

func (r *BillRepository) Update(bill *models.Bill) error {
	return r.db.Save(bill).Error
}

func (r *BillRepository) DeleteByReservationID(tx *gorm.DB, reservationID string) error {
	// Hard delete line items first, then the bill (soft delete would leave the unique index conflict)
	if err := tx.Unscoped().Where("bill_id IN (SELECT id FROM bills WHERE reservation_id = ?)", reservationID).Delete(&models.BillLineItem{}).Error; err != nil {
		return err
	}
	return tx.Unscoped().Where("reservation_id = ?", reservationID).Delete(&models.Bill{}).Error
}
