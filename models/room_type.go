package models

type RoomType struct {
	BaseModel
	HotelID     string `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Name        string `gorm:"not null;size:50" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (RoomType) TableName() string {
	return "room_types"
}
