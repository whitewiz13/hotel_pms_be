package service

import (
	"errors"
	"time"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

type ReservationService struct {
	db              *gorm.DB
	reservationRepo *repository.ReservationRepository
	roomRepo        *repository.RoomRepository
	billRepo        *repository.BillRepository
}

func NewReservationService(db *gorm.DB, reservationRepo *repository.ReservationRepository, roomRepo *repository.RoomRepository, billRepo *repository.BillRepository) *ReservationService {
	return &ReservationService{
		db:              db,
		reservationRepo: reservationRepo,
		roomRepo:        roomRepo,
		billRepo:        billRepo,
	}
}

// GetAvailability returns rooms available for the entire date range.
func (s *ReservationService) GetAvailability(hotelID string, query dto.AvailabilityQuery) ([]models.Room, error) {
	checkIn, err := time.Parse(dateLayout, query.CheckInDate)
	if err != nil {
		return nil, errors.New("invalid check_in date format, use YYYY-MM-DD")
	}
	checkOut, err := time.Parse(dateLayout, query.CheckOutDate)
	if err != nil {
		return nil, errors.New("invalid check_out date format, use YYYY-MM-DD")
	}

	if err := validateDateRange(checkIn, checkOut); err != nil {
		return nil, err
	}

	return s.reservationRepo.FindAvailableRooms(hotelID, checkIn, checkOut)
}

// Create creates a new reservation with inventory locking to prevent double booking.
func (s *ReservationService) Create(hotelID string, req dto.CreateReservationRequest) (*models.Reservation, error) {
	checkIn, err := time.Parse(dateLayout, req.CheckInDate)
	if err != nil {
		return nil, errors.New("invalid check_in_date format, use YYYY-MM-DD")
	}
	checkOut, err := time.Parse(dateLayout, req.CheckOutDate)
	if err != nil {
		return nil, errors.New("invalid check_out_date format, use YYYY-MM-DD")
	}

	if err := validateDateRange(checkIn, checkOut); err != nil {
		return nil, err
	}

	// Verify room belongs to hotel
	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil {
		return nil, errors.New("room not found")
	}
	if room.HotelID != hotelID {
		return nil, errors.New("room not found")
	}
	if !room.IsActive {
		return nil, errors.New("room is not active")
	}
	if room.Status == models.RoomStatusDirty || room.Status == models.RoomStatusCleaning || room.Status == models.RoomStatusMaintenance {
		return nil, errors.New("room is not available for booking")
	}

	reservation := &models.Reservation{
		HotelID:      hotelID,
		RoomID:       req.RoomID,
		GuestName:    req.GuestName,
		GuestPhone:   req.GuestPhone,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		Status:       models.ReservationStatusReserved,
		Notes:        req.Notes,
	}

	// Use transaction with row-level locking to prevent double booking
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Lock and check for conflicting inventory
		conflictCount, err := s.reservationRepo.LockInventoryForUpdate(tx, req.RoomID, hotelID, checkIn, checkOut)
		if err != nil {
			return errors.New("failed to check availability")
		}
		if conflictCount > 0 {
			return errors.New("room is not available for the selected dates")
		}

		// Create reservation
		if err := tx.Create(reservation).Error; err != nil {
			return errors.New("failed to create reservation")
		}

		// Mark inventory as booked
		resID := reservation.ID.String()
		if err := s.reservationRepo.CreateInventoryRecords(tx, hotelID, req.RoomID, resID, checkIn, checkOut); err != nil {
			return errors.New("failed to reserve inventory")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.reservationRepo.FindByID(reservation.ID.String())
}

// GetByID returns a single reservation.
func (s *ReservationService) GetByID(id, hotelID string) (*models.Reservation, error) {
	reservation, err := s.reservationRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}
	return reservation, nil
}

// List returns filtered reservations for a hotel.
func (s *ReservationService) List(hotelID string, query dto.ListReservationsQuery) ([]models.Reservation, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var dateFrom, dateTo, checkInDate, checkOutDate *time.Time
	if query.DateFrom != "" {
		t, err := time.Parse(dateLayout, query.DateFrom)
		if err != nil {
			return nil, 0, errors.New("invalid date_from format, use YYYY-MM-DD")
		}
		dateFrom = &t
	}
	if query.DateTo != "" {
		t, err := time.Parse(dateLayout, query.DateTo)
		if err != nil {
			return nil, 0, errors.New("invalid date_to format, use YYYY-MM-DD")
		}
		dateTo = &t
	}
	if query.CheckInDate != "" {
		t, err := time.Parse(dateLayout, query.CheckInDate)
		if err != nil {
			return nil, 0, errors.New("invalid check_in_date format, use YYYY-MM-DD")
		}
		checkInDate = &t
	}
	if query.CheckOutDate != "" {
		t, err := time.Parse(dateLayout, query.CheckOutDate)
		if err != nil {
			return nil, 0, errors.New("invalid check_out_date format, use YYYY-MM-DD")
		}
		checkOutDate = &t
	}

	return s.reservationRepo.FindByHotelID(hotelID, query.Status, query.RoomID, dateFrom, dateTo, checkInDate, checkOutDate, page, perPage)
}

// CheckIn transitions a reservation from reserved → checked_in.
// Generates an access PIN for the guest and sets it on the room.
func (s *ReservationService) CheckIn(id, hotelID string) (*models.Reservation, string, error) {
	reservation, err := s.reservationRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, "", errors.New("reservation not found")
	}

	if reservation.Status != models.ReservationStatusReserved {
		return nil, "", errors.New("only reserved bookings can be checked in")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if reservation.CheckInDate.After(today) {
		return nil, "", errors.New("cannot check in before the check-in date")
	}

	pin, err := utils.GeneratePin(6)
	if err != nil {
		return nil, "", errors.New("failed to generate access pin")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		reservation.Status = models.ReservationStatusCheckedIn
		if err := tx.Save(reservation).Error; err != nil {
			return errors.New("failed to check in")
		}

		// Mark room as occupied and set guest access PIN
		if err := tx.Model(&models.Room{}).Where("id = ?", reservation.RoomID).
			Updates(map[string]interface{}{
				"status":     models.RoomStatusOccupied,
				"access_pin": pin,
			}).Error; err != nil {
			return errors.New("failed to update room status")
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return reservation, pin, nil
}

// CheckOut transitions a reservation from checked_in → checked_out and frees inventory.
func (s *ReservationService) CheckOut(id, hotelID string) (*models.Reservation, error) {
	reservation, err := s.reservationRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}

	if reservation.Status != models.ReservationStatusCheckedIn {
		return nil, errors.New("only checked-in bookings can be checked out")
	}

	// Block checkout if there is a pending (unpaid) bill
	existingBill, _ := s.billRepo.FindByReservationID(id)
	if existingBill != nil && existingBill.Status == models.BillStatusPending {
		return nil, errors.New("cannot check out: there is a pending bill for this reservation, please settle it first")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		reservation.Status = models.ReservationStatusCheckedOut
		if err := tx.Save(reservation).Error; err != nil {
			return errors.New("failed to check out")
		}

		// Free up inventory so the room can be rebooked
		if err := s.reservationRepo.FreeInventory(tx, reservation.ID.String()); err != nil {
			return errors.New("failed to release inventory")
		}

		// Mark room as dirty and clear guest access PIN
		if err := tx.Model(&models.Room{}).Where("id = ?", reservation.RoomID).
			Updates(map[string]interface{}{
				"status":     models.RoomStatusDirty,
				"access_pin": "",
			}).Error; err != nil {
			return errors.New("failed to update room status")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return reservation, nil
}

// Cancel transitions a reservation from reserved → cancelled and frees inventory.
func (s *ReservationService) Cancel(id, hotelID string) (*models.Reservation, error) {
	reservation, err := s.reservationRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}

	if reservation.Status != models.ReservationStatusReserved {
		return nil, errors.New("only reserved bookings can be cancelled")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		reservation.Status = models.ReservationStatusCancelled
		if err := tx.Save(reservation).Error; err != nil {
			return errors.New("failed to cancel reservation")
		}

		// Free up inventory
		if err := s.reservationRepo.FreeInventory(tx, reservation.ID.String()); err != nil {
			return errors.New("failed to release inventory")
		}

		// Mark room as available again
		if err := tx.Model(&models.Room{}).Where("id = ?", reservation.RoomID).Update("status", models.RoomStatusAvailable).Error; err != nil {
			return errors.New("failed to update room status")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return reservation, nil
}

func validateDateRange(checkIn, checkOut time.Time) error {
	today := time.Now().Truncate(24 * time.Hour)
	if checkIn.Before(today) {
		return errors.New("check-in date cannot be in the past")
	}
	if !checkOut.After(checkIn) {
		return errors.New("check-out date must be after check-in date")
	}
	return nil
}
