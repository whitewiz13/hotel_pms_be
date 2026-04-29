package models

import (
	"time"

	"github.com/google/uuid"
)

type ReservationStatus string

const (
	ReservationStatusReserved   ReservationStatus = "reserved"
	ReservationStatusCheckedIn  ReservationStatus = "checked_in"
	ReservationStatusCheckedOut ReservationStatus = "checked_out"
	ReservationStatusCancelled  ReservationStatus = "cancelled"
)

type Reservation struct {
	BaseModel
	HotelID      string            `gorm:"type:uuid;not null;index" json:"hotel_id"`
	RoomID       string            `gorm:"type:uuid;not null;index" json:"room_id"`
	GuestID      string            `gorm:"type:uuid;not null;index" json:"guest_id"`
	GuestName    string            `gorm:"not null;size:255" json:"guest_name"`
	GuestPhone   string            `gorm:"size:20" json:"guest_phone"`
	GuestEmail   string            `gorm:"size:255" json:"guest_email"`
	CheckInDate  time.Time         `gorm:"type:date;not null;index" json:"check_in_date"`
	CheckOutDate time.Time         `gorm:"type:date;not null" json:"check_out_date"`
	Status       ReservationStatus `gorm:"type:varchar(20);not null;default:'reserved'" json:"status"`
	Notes        string            `gorm:"type:text" json:"notes"`

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
	Room  Room  `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Guest Guest `gorm:"foreignKey:GuestID" json:"guest,omitempty"`
}

func (Reservation) TableName() string {
	return "reservations"
}

// RoomInventory tracks availability per room per date.
// This is the source of truth for room availability.
type RoomInventory struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HotelID       string     `gorm:"type:uuid;not null;index" json:"hotel_id"`
	RoomID        string     `gorm:"type:uuid;not null" json:"room_id"`
	Date          time.Time  `gorm:"type:date;not null" json:"date"`
	IsAvailable   bool       `gorm:"not null;default:true" json:"is_available"`
	ReservationID *string    `gorm:"type:uuid" json:"reservation_id,omitempty"`

	Hotel       Hotel        `gorm:"foreignKey:HotelID" json:"-"`
	Room        Room         `gorm:"foreignKey:RoomID" json:"-"`
	Reservation *Reservation `gorm:"foreignKey:ReservationID" json:"-"`
}

func (RoomInventory) TableName() string {
	return "room_inventories"
}
