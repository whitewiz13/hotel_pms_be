package models

type Guest struct {
	BaseModel
	HotelID string `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Name    string `gorm:"not null;size:255" json:"name"`
	Email   string `gorm:"size:255" json:"email"`
	Phone   string `gorm:"size:20" json:"phone"`

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (Guest) TableName() string {
	return "guests"
}
