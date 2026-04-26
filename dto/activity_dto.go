package dto

import "github.com/hotelpms/backend/models"

type CreateActivityRequest struct {
	Name        string                  `json:"name" binding:"required,min=2"`
	Description string                  `json:"description"`
	Price       float64                 `json:"price" binding:"required,gt=0"`
	Category    models.ActivityCategory `json:"category" binding:"required,oneof=cab spa tour laundry other"`
}

type UpdateActivityRequest struct {
	Name        *string                  `json:"name" binding:"omitempty,min=2"`
	Description *string                  `json:"description"`
	Price       *float64                 `json:"price" binding:"omitempty,gt=0"`
	Category    *models.ActivityCategory `json:"category" binding:"omitempty,oneof=cab spa tour laundry other"`
	IsAvailable *bool                    `json:"is_available"`
}

type ListActivitiesQuery struct {
	Category string `form:"category"`
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
}

type CreateActivityBookingRequest struct {
	RoomID        string `json:"room_id" binding:"required,uuid"`
	ReservationID string `json:"reservation_id" binding:"required,uuid"`
	ActivityID    string `json:"activity_id" binding:"required,uuid"`
	GuestName     string `json:"guest_name" binding:"required,min=2"`
	GuestID       string `json:"guest_id"`
	ScheduledAt   string `json:"scheduled_at"`
	Notes         string `json:"notes"`
}

type UpdateActivityBookingStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=confirmed completed cancelled"`
}

type ListActivityBookingsQuery struct {
	Status        string `form:"status"`
	RoomID        string `form:"room_id"`
	ReservationID string `form:"reservation_id"`
	ActivityID    string `form:"activity_id"`
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
}
