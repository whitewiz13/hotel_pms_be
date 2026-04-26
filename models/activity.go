package models

import "time"

type ActivityCategory string

const (
	ActivityCategoryCab     ActivityCategory = "cab"
	ActivityCategorySpa     ActivityCategory = "spa"
	ActivityCategoryTour    ActivityCategory = "tour"
	ActivityCategoryLaundry ActivityCategory = "laundry"
	ActivityCategoryOther   ActivityCategory = "other"
)

type Activity struct {
	BaseModel
	HotelID     string           `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Name        string           `gorm:"not null;size:255" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Price       float64          `gorm:"not null;default:0" json:"price"`
	Category    ActivityCategory `gorm:"type:varchar(30);not null;default:'other'" json:"category"`
	IsAvailable bool             `gorm:"default:true" json:"is_available"`

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (Activity) TableName() string {
	return "activities"
}

type ActivityBookingStatus string

const (
	ActivityBookingPending   ActivityBookingStatus = "pending"
	ActivityBookingConfirmed ActivityBookingStatus = "confirmed"
	ActivityBookingCompleted ActivityBookingStatus = "completed"
	ActivityBookingCancelled ActivityBookingStatus = "cancelled"
)

type ActivityBooking struct {
	BaseModel
	HotelID       string                `gorm:"type:uuid;not null;index" json:"hotel_id"`
	RoomID        string                `gorm:"type:uuid;not null;index" json:"room_id"`
	ReservationID string                `gorm:"type:uuid;not null;index" json:"reservation_id"`
	ActivityID    string                `gorm:"type:uuid;not null;index" json:"activity_id"`
	GuestID       string                `gorm:"type:uuid;not null;index" json:"guest_id"`
	GuestName     string                `gorm:"not null;size:255" json:"guest_name"`
	ScheduledAt   *time.Time            `gorm:"type:timestamptz" json:"scheduled_at,omitempty"`
	Status        ActivityBookingStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Amount        float64               `gorm:"not null;default:0" json:"amount"`
	Notes         string                `gorm:"type:text" json:"notes"`

	Hotel       Hotel       `gorm:"foreignKey:HotelID" json:"-"`
	Room        Room        `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Reservation Reservation `gorm:"foreignKey:ReservationID" json:"-"`
	Activity    Activity    `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	Guest       Guest       `gorm:"foreignKey:GuestID" json:"guest,omitempty"`
}

func (ActivityBooking) TableName() string {
	return "activity_bookings"
}
