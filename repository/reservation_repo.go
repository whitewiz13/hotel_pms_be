package repository

import (
	"time"

	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(reservation *models.Reservation) error {
	return r.db.Create(reservation).Error
}

func (r *ReservationRepository) FindByID(id string) (*models.Reservation, error) {
	var reservation models.Reservation
	err := r.db.Preload("Room").Where("id = ?", id).First(&reservation).Error
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) FindByIDAndHotel(id, hotelID string) (*models.Reservation, error) {
	var reservation models.Reservation
	err := r.db.Preload("Room").Where("id = ? AND hotel_id = ?", id, hotelID).First(&reservation).Error
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) FindByHotelID(hotelID string, status, roomID string, dateFrom, dateTo, checkInDate, checkOutDate *time.Time, page, perPage int) ([]models.Reservation, int64, error) {
	var reservations []models.Reservation
	var total int64

	query := r.db.Where("reservations.hotel_id = ?", hotelID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if roomID != "" {
		query = query.Where("room_id = ?", roomID)
	}
	if dateFrom != nil {
		query = query.Where("check_in_date >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("check_out_date <= ?", *dateTo)
	}
	if checkInDate != nil {
		query = query.Where("check_in_date = ?", *checkInDate)
	}
	if checkOutDate != nil {
		query = query.Where("check_out_date = ?", *checkOutDate)
	}

	query.Model(&models.Reservation{}).Count(&total)

	err := query.Preload("Room").
		Order("check_in_date DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&reservations).Error

	return reservations, total, err
}

func (r *ReservationRepository) Update(reservation *models.Reservation) error {
	return r.db.Save(reservation).Error
}

// FindAvailableRooms returns rooms that are available for ALL dates in the range [checkIn, checkOut).
// A room is available if no inventory record blocks it for any date in the range.
func (r *ReservationRepository) FindAvailableRooms(hotelID string, checkIn, checkOut time.Time) ([]models.Room, error) {
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return nil, nil
	}

	var rooms []models.Room
	// A room is available if it has no inventory record marked as unavailable
	// for any date in the range. We look for rooms that do NOT have a blocking inventory row.
	err := r.db.
		Where("hotel_id = ? AND is_active = true", hotelID).
		Where("id NOT IN (?)",
			r.db.Model(&models.RoomInventory{}).
				Select("room_id").
				Where("hotel_id = ? AND date >= ? AND date < ? AND is_available = false", hotelID, checkIn, checkOut),
		).
		Preload("Amenities").
		Find(&rooms).Error

	return rooms, err
}

// LockInventoryForUpdate locks inventory rows for the given room and date range using SELECT FOR UPDATE.
// Returns the count of unavailable dates. Must be called within a transaction.
func (r *ReservationRepository) LockInventoryForUpdate(tx *gorm.DB, roomID, hotelID string, checkIn, checkOut time.Time) (int64, error) {
	var records []models.RoomInventory
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("room_id = ? AND hotel_id = ? AND date >= ? AND date < ? AND is_available = false",
			roomID, hotelID, checkIn, checkOut).
		Find(&records).Error
	return int64(len(records)), err
}

// CreateInventoryRecords marks inventory dates as booked for a reservation.
func (r *ReservationRepository) CreateInventoryRecords(tx *gorm.DB, hotelID, roomID, reservationID string, checkIn, checkOut time.Time) error {
	current := checkIn
	for current.Before(checkOut) {
		result := tx.Model(&models.RoomInventory{}).
			Where("room_id = ? AND hotel_id = ? AND date = ?", roomID, hotelID, current).
			Updates(map[string]interface{}{
				"is_available":   false,
				"reservation_id": reservationID,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// No existing record — create one via raw SQL to avoid GORM zero-value issue with is_available
			if err := tx.Exec(
				`INSERT INTO room_inventories (hotel_id, room_id, date, is_available, reservation_id) VALUES (?, ?, ?, false, ?)`,
				hotelID, roomID, current, reservationID,
			).Error; err != nil {
				return err
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return nil
}

// FreeInventory marks inventory records as available again for a cancelled reservation.
func (r *ReservationRepository) FreeInventory(tx *gorm.DB, reservationID string) error {
	return tx.Model(&models.RoomInventory{}).
		Where("reservation_id = ?", reservationID).
		Updates(map[string]interface{}{
			"is_available":   true,
			"reservation_id": nil,
		}).Error
}
