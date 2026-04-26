package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type GuestService struct {
	reservationRepo  *repository.ReservationRepository
	orderService     *OrderService
	activityService  *ActivityService
	menuService      *MenuService
	orderRepo        *repository.OrderRepository
	activityRepo     *repository.ActivityRepository
	settingsRepo     *repository.GuestSettingsRepository
}

func NewGuestService(
	reservationRepo *repository.ReservationRepository,
	orderService *OrderService,
	activityService *ActivityService,
	menuService *MenuService,
	orderRepo *repository.OrderRepository,
	activityRepo *repository.ActivityRepository,
	settingsRepo *repository.GuestSettingsRepository,
) *GuestService {
	return &GuestService{
		reservationRepo:  reservationRepo,
		orderService:     orderService,
		activityService:  activityService,
		menuService:      menuService,
		orderRepo:        orderRepo,
		activityRepo:     activityRepo,
		settingsRepo:     settingsRepo,
	}
}

// getActiveReservation resolves the checked-in reservation for a guest's room.
func (s *GuestService) getActiveReservation(roomID, hotelID string) (*models.Reservation, error) {
	reservation, err := s.reservationRepo.FindCheckedInByRoomID(roomID, hotelID)
	if err != nil {
		return nil, errors.New("no active reservation found for this room")
	}
	return reservation, nil
}

func (s *GuestService) GetMyReservation(roomID, hotelID string) (*models.Reservation, error) {
	return s.getActiveReservation(roomID, hotelID)
}

func (s *GuestService) ListMenu(hotelID string, page, perPage int) ([]models.MenuItem, int64, error) {
	query := dto.ListMenuQuery{
		Page:    page,
		PerPage: perPage,
	}
	return s.menuService.List(hotelID, query)
}

func (s *GuestService) ListActivities(hotelID string, page, perPage int) ([]models.Activity, int64, error) {
	query := dto.ListActivitiesQuery{
		Page:    page,
		PerPage: perPage,
	}
	return s.activityService.List(hotelID, query)
}

func (s *GuestService) PlaceOrder(hotelID, roomID string, req dto.GuestPlaceOrderRequest) (*models.Order, error) {
	// Check if orders are allowed
	settings, _ := s.settingsRepo.FindByHotelID(hotelID)
	if settings != nil && !settings.AllowOrders {
		return nil, errors.New("ordering is currently disabled by the hotel")
	}

	reservation, err := s.getActiveReservation(roomID, hotelID)
	if err != nil {
		return nil, err
	}

	createReq := dto.CreateOrderRequest{
		RoomID:        roomID,
		ReservationID: reservation.ID.String(),
		GuestName:     reservation.GuestName,
		GuestID:       reservation.GuestID,
		Items:         req.Items,
		Notes:         req.Notes,
	}

	return s.orderService.Create(hotelID, createReq)
}

func (s *GuestService) BookActivity(hotelID, roomID string, req dto.GuestBookActivityRequest) (*models.ActivityBooking, error) {
	// Check if activities are allowed
	settings, _ := s.settingsRepo.FindByHotelID(hotelID)
	if settings != nil && !settings.AllowActivities {
		return nil, errors.New("activity booking is currently disabled by the hotel")
	}

	reservation, err := s.getActiveReservation(roomID, hotelID)
	if err != nil {
		return nil, err
	}

	createReq := dto.CreateActivityBookingRequest{
		RoomID:        roomID,
		ReservationID: reservation.ID.String(),
		ActivityID:    req.ActivityID,
		GuestName:     reservation.GuestName,
		GuestID:       reservation.GuestID,
		ScheduledAt:   req.ScheduledAt,
		Notes:         req.Notes,
	}

	return s.activityService.CreateBooking(hotelID, createReq)
}

func (s *GuestService) ListMyOrders(hotelID, roomID string, page, perPage int) ([]models.Order, int64, error) {
	reservation, err := s.getActiveReservation(roomID, hotelID)
	if err != nil {
		return nil, 0, err
	}

	query := dto.ListOrdersQuery{
		ReservationID: reservation.ID.String(),
		Page:          page,
		PerPage:       perPage,
	}

	return s.orderService.List(hotelID, query)
}

func (s *GuestService) ListMyActivityBookings(hotelID, roomID string, page, perPage int) ([]models.ActivityBooking, int64, error) {
	reservation, err := s.getActiveReservation(roomID, hotelID)
	if err != nil {
		return nil, 0, err
	}

	query := dto.ListActivityBookingsQuery{
		ReservationID: reservation.ID.String(),
		Page:          page,
		PerPage:       perPage,
	}

	return s.activityService.ListBookings(hotelID, query)
}

func (s *GuestService) GetDashboard(hotelID, roomID string) (*dto.GuestDashboardResponse, error) {
	reservation, err := s.getActiveReservation(roomID, hotelID)
	if err != nil {
		return nil, err
	}

	resID := reservation.ID.String()

	// Order stats
	orderStats, err := s.orderRepo.CountByReservationGroupedByStatus(resID)
	if err != nil {
		orderStats = make(map[string]int64)
	}
	var totalOrders int64
	for _, c := range orderStats {
		totalOrders += c
	}

	// Order spend (non-cancelled orders)
	var orderSpend float64
	orders, err := s.orderRepo.FindByReservationID(resID)
	if err == nil {
		for _, o := range orders {
			orderSpend += o.TotalAmount
		}
	}

	// Activity stats
	activityStats, err := s.activityRepo.CountBookingsByReservationGroupedByStatus(resID)
	if err != nil {
		activityStats = make(map[string]int64)
	}
	var totalActivities int64
	for _, c := range activityStats {
		totalActivities += c
	}

	// Activity spend (non-cancelled bookings)
	var activitySpend float64
	bookings, err := s.activityRepo.FindBookingsByReservationID(resID)
	if err == nil {
		for _, b := range bookings {
			activitySpend += b.Amount
		}
	}

	return &dto.GuestDashboardResponse{
		RoomNumber:      reservation.Room.RoomNumber,
		GuestName:       reservation.GuestName,
		CheckInDate:     reservation.CheckInDate.Format("2006-01-02"),
		CheckOutDate:    reservation.CheckOutDate.Format("2006-01-02"),
		OrderStats:      orderStats,
		TotalOrders:     totalOrders,
		OrderSpend:      orderSpend,
		ActivityStats:   activityStats,
		TotalActivities: totalActivities,
		ActivitySpend:   activitySpend,
	}, nil
}
