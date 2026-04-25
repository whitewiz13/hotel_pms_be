package service

import (
	"errors"
	"time"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type ActivityService struct {
	activityRepo    *repository.ActivityRepository
	roomRepo        *repository.RoomRepository
	reservationRepo *repository.ReservationRepository
}

func NewActivityService(
	activityRepo *repository.ActivityRepository,
	roomRepo *repository.RoomRepository,
	reservationRepo *repository.ReservationRepository,
) *ActivityService {
	return &ActivityService{
		activityRepo:    activityRepo,
		roomRepo:        roomRepo,
		reservationRepo: reservationRepo,
	}
}

// --- Activity CRUD ---

func (s *ActivityService) Create(hotelID string, req dto.CreateActivityRequest) (*models.Activity, error) {
	activity := &models.Activity{
		HotelID:     hotelID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		IsAvailable: true,
	}

	if err := s.activityRepo.Create(activity); err != nil {
		return nil, errors.New("failed to create activity")
	}

	return activity, nil
}

func (s *ActivityService) GetByID(id, hotelID string) (*models.Activity, error) {
	activity, err := s.activityRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("activity not found")
	}
	return activity, nil
}

func (s *ActivityService) List(hotelID string, query dto.ListActivitiesQuery) ([]models.Activity, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.activityRepo.FindByHotelID(hotelID, query.Category, page, perPage)
}

func (s *ActivityService) Update(id, hotelID string, req dto.UpdateActivityRequest) (*models.Activity, error) {
	activity, err := s.activityRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("activity not found")
	}

	if req.Name != nil {
		activity.Name = *req.Name
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.Price != nil {
		activity.Price = *req.Price
	}
	if req.Category != nil {
		activity.Category = *req.Category
	}
	if req.IsAvailable != nil {
		activity.IsAvailable = *req.IsAvailable
	}

	if err := s.activityRepo.Update(activity); err != nil {
		return nil, errors.New("failed to update activity")
	}

	return activity, nil
}

func (s *ActivityService) Delete(id, hotelID string) error {
	_, err := s.activityRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return errors.New("activity not found")
	}

	return s.activityRepo.Delete(id)
}

// --- Activity Booking ---

func (s *ActivityService) CreateBooking(hotelID string, req dto.CreateActivityBookingRequest) (*models.ActivityBooking, error) {
	// Verify activity exists and is available
	activity, err := s.activityRepo.FindByIDAndHotel(req.ActivityID, hotelID)
	if err != nil {
		return nil, errors.New("activity not found")
	}
	if !activity.IsAvailable {
		return nil, errors.New("activity is not available")
	}

	// Verify room belongs to hotel
	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil || room.HotelID != hotelID {
		return nil, errors.New("room not found")
	}

	// Verify reservation is checked in
	reservation, err := s.reservationRepo.FindByIDAndHotel(req.ReservationID, hotelID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}
	if reservation.Status != models.ReservationStatusCheckedIn {
		return nil, errors.New("activities can only be booked for checked-in reservations")
	}
	if reservation.RoomID != req.RoomID {
		return nil, errors.New("reservation does not match the specified room")
	}

	booking := &models.ActivityBooking{
		HotelID:       hotelID,
		RoomID:        req.RoomID,
		ReservationID: req.ReservationID,
		ActivityID:    req.ActivityID,
		GuestName:     req.GuestName,
		Status:        models.ActivityBookingPending,
		Amount:        activity.Price,
		Notes:         req.Notes,
	}

	if req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			return nil, errors.New("invalid scheduled_at format, use RFC3339 (e.g. 2026-04-25T14:00:00Z)")
		}
		booking.ScheduledAt = &t
	}

	if err := s.activityRepo.CreateBooking(booking); err != nil {
		return nil, errors.New("failed to create activity booking")
	}

	return s.activityRepo.FindBookingByID(booking.ID.String())
}

func (s *ActivityService) GetBookingByID(id, hotelID string) (*models.ActivityBooking, error) {
	booking, err := s.activityRepo.FindBookingByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("activity booking not found")
	}
	return booking, nil
}

func (s *ActivityService) ListBookings(hotelID string, query dto.ListActivityBookingsQuery) ([]models.ActivityBooking, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.activityRepo.FindBookingsByHotelID(hotelID, query.Status, query.RoomID, query.ReservationID, query.ActivityID, page, perPage)
}

func (s *ActivityService) UpdateBookingStatus(id, hotelID string, req dto.UpdateActivityBookingStatusRequest) (*models.ActivityBooking, error) {
	booking, err := s.activityRepo.FindBookingByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("activity booking not found")
	}

	newStatus := models.ActivityBookingStatus(req.Status)

	// Validate transitions
	switch booking.Status {
	case models.ActivityBookingPending:
		if newStatus != models.ActivityBookingConfirmed && newStatus != models.ActivityBookingCancelled {
			return nil, errors.New("pending bookings can only move to confirmed or cancelled")
		}
	case models.ActivityBookingConfirmed:
		if newStatus != models.ActivityBookingCompleted && newStatus != models.ActivityBookingCancelled {
			return nil, errors.New("confirmed bookings can only move to completed or cancelled")
		}
	default:
		return nil, errors.New("booking cannot be updated from current status")
	}

	booking.Status = newStatus
	if err := s.activityRepo.UpdateBooking(booking); err != nil {
		return nil, errors.New("failed to update activity booking")
	}

	return booking, nil
}
