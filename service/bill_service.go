package service

import (
	"errors"
	"fmt"
	"math"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"gorm.io/gorm"
)

type BillService struct {
	db                  *gorm.DB
	billRepo            *repository.BillRepository
	reservationRepo     *repository.ReservationRepository
	orderRepo           *repository.OrderRepository
	activityRepo        *repository.ActivityRepository
	roomRepo            *repository.RoomRepository
	notificationService *NotificationService
}

func NewBillService(
	db *gorm.DB,
	billRepo *repository.BillRepository,
	reservationRepo *repository.ReservationRepository,
	orderRepo *repository.OrderRepository,
	activityRepo *repository.ActivityRepository,
	roomRepo *repository.RoomRepository,
	notificationService *NotificationService,
) *BillService {
	return &BillService{
		db:                  db,
		billRepo:            billRepo,
		reservationRepo:     reservationRepo,
		orderRepo:           orderRepo,
		activityRepo:        activityRepo,
		roomRepo:            roomRepo,
		notificationService: notificationService,
	}
}

// Generate creates a bill for a reservation, calculating all charges.
func (s *BillService) Generate(hotelID, reservationID string, req dto.GenerateBillRequest) (*models.Bill, error) {
	// If a bill already exists for this reservation, delete it so we can regenerate
	existing, _ := s.billRepo.FindByReservationID(reservationID)
	if existing != nil {
		if existing.Status == models.BillStatusPaid {
			return nil, errors.New("cannot regenerate a bill that is already paid")
		}
		if err := s.billRepo.DeleteByReservationID(s.db, reservationID); err != nil {
			return nil, errors.New("failed to remove existing bill for regeneration")
		}
	}

	// Get reservation
	reservation, err := s.reservationRepo.FindByIDAndHotel(reservationID, hotelID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}

	if reservation.Status != models.ReservationStatusCheckedIn && reservation.Status != models.ReservationStatusCheckedOut {
		return nil, errors.New("bill can only be generated for checked-in or checked-out reservations")
	}

	// Get room for pricing
	room, err := s.roomRepo.FindByID(reservation.RoomID)
	if err != nil {
		return nil, errors.New("room not found")
	}

	// Calculate nights
	nights := int(reservation.CheckOutDate.Sub(reservation.CheckInDate).Hours() / 24)
	if nights < 1 {
		nights = 1
	}

	var lineItems []models.BillLineItem

	// 1. Room charges
	roomCharges := room.PricePerNight * float64(nights)
	lineItems = append(lineItems, models.BillLineItem{
		Type:        models.BillLineRoom,
		Description: fmt.Sprintf("Room %s - %s (%d night(s) × $%.2f)", room.RoomNumber, room.RoomType, nights, room.PricePerNight),
		Amount:      roomCharges,
	})

	// 2. Room service orders
	var roomServiceTotal float64
	orders, err := s.orderRepo.FindByReservationID(reservationID)
	if err == nil {
		for _, order := range orders {
			roomServiceTotal += order.TotalAmount
			orderID := order.ID.String()
			lineItems = append(lineItems, models.BillLineItem{
				Type:        models.BillLineRoomService,
				Description: fmt.Sprintf("Room Service Order #%s", order.ID.String()[:8]),
				Amount:      order.TotalAmount,
				ReferenceID: &orderID,
			})
		}
	}

	// 3. Activity bookings
	var activityTotal float64
	bookings, err := s.activityRepo.FindBookingsByReservationID(reservationID)
	if err == nil {
		for _, booking := range bookings {
			activityTotal += booking.Amount
			bookingID := booking.ID.String()
			lineItems = append(lineItems, models.BillLineItem{
				Type:        models.BillLineActivity,
				Description: fmt.Sprintf("Activity: %s", booking.Activity.Name),
				Amount:      booking.Amount,
				ReferenceID: &bookingID,
			})
		}
	}

	// 4. Upfront payment (negative line item)
	upfrontPaid := req.UpfrontPaid
	if upfrontPaid > 0 {
		lineItems = append(lineItems, models.BillLineItem{
			Type:        models.BillLineUpfront,
			Description: "Upfront Payment",
			Amount:      -upfrontPaid,
		})
	}

	// Calculate totals
	subtotal := roomCharges + roomServiceTotal + activityTotal - upfrontPaid
	taxRate := req.TaxRate
	taxAmount := math.Round(subtotal*taxRate) / 100 // tax_rate is percentage
	totalAmount := subtotal + taxAmount

	bill := &models.Bill{
		HotelID:          hotelID,
		ReservationID:    reservationID,
		RoomID:           reservation.RoomID,
		GuestID:          reservation.GuestID,
		GuestName:        reservation.GuestName,
		RoomCharges:      roomCharges,
		UpfrontPaid:      upfrontPaid,
		RoomServiceTotal: roomServiceTotal,
		ActivityTotal:    activityTotal,
		Subtotal:         subtotal,
		TaxRate:          taxRate,
		TaxAmount:        taxAmount,
		TotalAmount:      totalAmount,
		Status:           models.BillStatusPending,
		Notes:            req.Notes,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(bill).Error; err != nil {
			return errors.New("failed to create bill")
		}

		// Set bill ID on all line items
		for i := range lineItems {
			lineItems[i].BillID = bill.ID.String()
		}

		if len(lineItems) > 0 {
			if err := tx.Create(&lineItems).Error; err != nil {
				return errors.New("failed to create bill line items")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Notify front-desk about the generated bill
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"Bill Generated",
			fmt.Sprintf("Bill for %s — $%.2f (Room %s)", reservation.GuestName, totalAmount, room.RoomNumber),
			map[string]string{
				"type":           "bill_generated",
				"bill_id":        bill.ID.String(),
				"reservation_id": reservationID,
				"hotel_id":       hotelID,
			},
			"billing:read", "billing:pay",
		)
	}

	return s.billRepo.FindByID(bill.ID.String())
}

func (s *BillService) GetByID(id, hotelID string) (*models.Bill, error) {
	bill, err := s.billRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("bill not found")
	}
	return bill, nil
}

func (s *BillService) GetByReservationID(reservationID, hotelID string) (*models.Bill, error) {
	bill, err := s.billRepo.FindByReservationID(reservationID)
	if err != nil {
		return nil, errors.New("bill not found for this reservation")
	}
	if bill.HotelID != hotelID {
		return nil, errors.New("bill not found")
	}
	return bill, nil
}

func (s *BillService) List(hotelID string, query dto.ListBillsQuery) ([]models.Bill, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.billRepo.FindByHotelID(hotelID, query.Status, query.ReservationID, page, perPage)
}

func (s *BillService) MarkPaid(id, hotelID string) (*models.Bill, error) {
	bill, err := s.billRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("bill not found")
	}

	if bill.Status == models.BillStatusPaid {
		return nil, errors.New("bill is already paid")
	}

	bill.Status = models.BillStatusPaid
	if err := s.billRepo.Update(bill); err != nil {
		return nil, errors.New("failed to update bill status")
	}

	// Notify staff that bill is paid — guest can now check out
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"Bill Paid",
			fmt.Sprintf("%s paid $%.2f — ready for checkout", bill.GuestName, bill.TotalAmount),
			map[string]string{
				"type":    "bill_paid",
				"bill_id": id,
				"hotel_id": hotelID,
			},
			"billing:read", "reservations:check_out",
		)
	}

	return bill, nil
}
