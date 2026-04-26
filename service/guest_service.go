package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type GuestService struct {
	reservationRepo *repository.ReservationRepository
	orderService    *OrderService
	activityService *ActivityService
	menuService     *MenuService
	orderRepo       *repository.OrderRepository
	activityRepo    *repository.ActivityRepository
}

func NewGuestService(
	reservationRepo *repository.ReservationRepository,
	orderService *OrderService,
	activityService *ActivityService,
	menuService *MenuService,
	orderRepo *repository.OrderRepository,
	activityRepo *repository.ActivityRepository,
) *GuestService {
	return &GuestService{
		reservationRepo: reservationRepo,
		orderService:    orderService,
		activityService: activityService,
		menuService:     menuService,
		orderRepo:       orderRepo,
		activityRepo:    activityRepo,
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
