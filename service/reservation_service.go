package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

type ReservationService struct {
	db                  *gorm.DB
	reservationRepo     *repository.ReservationRepository
	roomRepo            *repository.RoomRepository
	billRepo            *repository.BillRepository
	guestRepo           *repository.GuestRepository
	notificationService *NotificationService
	planService         *PlanService
}

func NewReservationService(db *gorm.DB, reservationRepo *repository.ReservationRepository, roomRepo *repository.RoomRepository, billRepo *repository.BillRepository, guestRepo *repository.GuestRepository, notificationService *NotificationService, planService *PlanService) *ReservationService {
	return &ReservationService{
		db:                  db,
		reservationRepo:     reservationRepo,
		roomRepo:            roomRepo,
		billRepo:            billRepo,
		guestRepo:           guestRepo,
		notificationService: notificationService,
		planService:         planService,
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
func (s *ReservationService) Create(hotelID string, userID string, req dto.CreateReservationRequest) (*models.Reservation, error) {
	// Check plan reservation limit
	if err := s.planService.CheckReservationLimit(hotelID); err != nil {
		return nil, err
	}

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
		GuestEmail:   req.GuestEmail,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		Status:       models.ReservationStatusReserved,
		Notes:        req.Notes,
		CreatedByID:  userID,
	}

	// Use transaction with row-level locking to prevent double booking
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Find or create guest record
		guest, err := s.guestRepo.FindOrCreate(tx, hotelID, req.GuestName, req.GuestPhone, req.GuestEmail)
		if err != nil {
			return errors.New("failed to create guest record")
		}
		reservation.GuestID = guest.ID.String()

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

	// Notify front-desk / reservations staff
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"New Reservation",
			fmt.Sprintf("%s — Room %s (%s to %s)", req.GuestName, room.RoomNumber, req.CheckInDate, req.CheckOutDate),
			map[string]string{
				"type":           "new_reservation",
				"reservation_id": reservation.ID.String(),
				"hotel_id":       hotelID,
			},
			"reservations:read", "reservations:check_in",
		)
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
// Optionally updates guest email and identity document details.
func (s *ReservationService) CheckIn(id, hotelID string, req dto.CheckInRequest) (*models.Reservation, string, error) {
	reservation, err := s.reservationRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, "", errors.New("reservation not found")
	}

	if reservation.Status != models.ReservationStatusReserved {
		return nil, "", errors.New("only reserved bookings can be checked in")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	checkInDay := time.Date(reservation.CheckInDate.Year(), reservation.CheckInDate.Month(), reservation.CheckInDate.Day(), 0, 0, 0, 0, time.UTC)
	if checkInDay.After(today) {
		return nil, "", errors.New("cannot check in before the check-in date")
	}

	pin, err := utils.GeneratePin(6)
	if err != nil {
		return nil, "", errors.New("failed to generate access pin")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update guest_email on reservation if provided at check-in
		if req.GuestEmail != nil && *req.GuestEmail != "" {
			reservation.GuestEmail = *req.GuestEmail
		}
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

		// Update guest email and identity details if provided
		guestUpdates := map[string]interface{}{}
		if req.GuestEmail != nil && *req.GuestEmail != "" {
			guestUpdates["email"] = *req.GuestEmail
		}
		if req.IDType != nil && *req.IDType != "" {
			guestUpdates["id_type"] = *req.IDType
		}
		if req.IDNumber != nil && *req.IDNumber != "" {
			guestUpdates["id_number"] = *req.IDNumber
		}
		if req.IDDocumentURL != nil && *req.IDDocumentURL != "" {
			guestUpdates["id_document_url"] = *req.IDDocumentURL
		}
		if len(guestUpdates) > 0 {
			if err := tx.Model(&models.Guest{}).Where("id = ?", reservation.GuestID).
				Updates(guestUpdates).Error; err != nil {
				return errors.New("failed to update guest details")
			}
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	// Notify staff about the check-in
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"Guest Checked In",
			fmt.Sprintf("%s checked into Room %s", reservation.GuestName, reservation.Room.RoomNumber),
			map[string]string{
				"type":           "check_in",
				"reservation_id": id,
				"hotel_id":       hotelID,
			},
			"reservations:read", "reservations:check_in",
		)
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

	// Notify staff about check-out; housekeeping will need to clean the room
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"Guest Checked Out",
			fmt.Sprintf("%s checked out of Room %s — room needs cleaning", reservation.GuestName, reservation.Room.RoomNumber),
			map[string]string{
				"type":           "check_out",
				"reservation_id": id,
				"hotel_id":       hotelID,
			},
			"reservations:read", "housekeeping:assign",
		)
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

	// Notify staff about cancellation
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"Reservation Cancelled",
			fmt.Sprintf("%s \u2014 Room %s is now available", reservation.GuestName, reservation.Room.RoomNumber),
			map[string]string{
				"type":           "reservation_cancelled",
				"reservation_id": id,
				"hotel_id":       hotelID,
			},
			"reservations:read",
		)
	}

	return reservation, nil
}

func validateDateRange(checkIn, checkOut time.Time) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	checkInDay := time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), 0, 0, 0, 0, time.UTC)
	if checkInDay.Before(today) {
		return errors.New("check-in date cannot be in the past")
	}
	if !checkOut.After(checkIn) {
		return errors.New("check-out date must be after check-in date")
	}
	return nil
}
