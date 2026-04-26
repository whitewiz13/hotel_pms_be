package dto

import "github.com/hotelpms/backend/models"

type CreateRoomRequest struct {
	RoomNumber    string             `json:"room_number" binding:"required"`
	RoomType      string             `json:"room_type" binding:"required"`
	Floor         int                `json:"floor" binding:"required,min=0"`
	PricePerNight float64            `json:"price_per_night" binding:"required,gt=0"`
	Description   string             `json:"description"`
	MaxOccupancy  int                `json:"max_occupancy" binding:"required,min=1"`
	AmenityIDs    []string           `json:"amenity_ids"`
}

type UpdateRoomRequest struct {
	RoomNumber    *string            `json:"room_number"`
	RoomType      *string            `json:"room_type"`
	Floor         *int               `json:"floor" binding:"omitempty,min=0"`
	Status        *models.RoomStatus `json:"status" binding:"omitempty,oneof=available occupied maintenance cleaning"`
	PricePerNight *float64           `json:"price_per_night" binding:"omitempty,gt=0"`
	Description   *string            `json:"description"`
	MaxOccupancy  *int               `json:"max_occupancy" binding:"omitempty,min=1"`
	IsActive      *bool              `json:"is_active"`
	AmenityIDs    []string           `json:"amenity_ids"`
}

type SetRoomPinRequest struct {
	Pin string `json:"pin" binding:"required,len=6,numeric"`
}
