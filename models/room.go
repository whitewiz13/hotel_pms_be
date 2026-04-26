package models

type RoomStatus string

const (
	RoomStatusAvailable   RoomStatus = "available"
	RoomStatusOccupied    RoomStatus = "occupied"
	RoomStatusDirty       RoomStatus = "dirty"
	RoomStatusCleaning    RoomStatus = "cleaning"
	RoomStatusMaintenance RoomStatus = "maintenance"
)

type Room struct {
	BaseModel
	HotelID      string     `gorm:"type:uuid;not null;index" json:"hotel_id"`
	RoomNumber   string     `gorm:"not null;size:20" json:"room_number"`
	RoomType     string     `gorm:"type:varchar(50);not null" json:"room_type"`
	Floor        int        `gorm:"not null;default:1" json:"floor"`
	Status       RoomStatus `gorm:"type:varchar(20);not null;default:'available'" json:"status"`
	PricePerNight float64   `gorm:"not null;default:0" json:"price_per_night"`
	Description  string     `gorm:"type:text" json:"description"`
	MaxOccupancy int        `gorm:"not null;default:2" json:"max_occupancy"`
	AccessPin    string     `gorm:"size:6" json:"access_pin,omitempty"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	Hotel     Hotel     `gorm:"foreignKey:HotelID" json:"-"`
	Amenities []Amenity `gorm:"many2many:room_amenities;" json:"amenities,omitempty"`
}

func (Room) TableName() string {
	return "rooms"
}
