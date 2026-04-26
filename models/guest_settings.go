package models

type GuestSettings struct {
	BaseModel
	HotelID         string `gorm:"type:uuid;not null;uniqueIndex" json:"hotel_id"`
	WifiPassword    string `gorm:"size:255" json:"wifi_password"`
	AllowOrders     bool   `gorm:"not null;default:true" json:"allow_orders"`
	AllowActivities bool   `gorm:"not null;default:true" json:"allow_activities"`

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (GuestSettings) TableName() string {
	return "guest_settings"
}
