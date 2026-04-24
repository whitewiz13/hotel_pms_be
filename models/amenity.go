package models

type AmenityCategory string

const (
	AmenityCategoryRoom     AmenityCategory = "room"
	AmenityCategoryBathroom AmenityCategory = "bathroom"
	AmenityCategoryGeneral  AmenityCategory = "general"
	AmenityCategoryDining   AmenityCategory = "dining"
	AmenityCategoryRecreation AmenityCategory = "recreation"
)

type Amenity struct {
	BaseModel
	Name        string          `gorm:"not null;size:255" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Icon        string          `gorm:"size:100" json:"icon"`
	Category    AmenityCategory `gorm:"type:varchar(30);not null;default:'general'" json:"category"`
	IsActive    bool            `gorm:"default:true" json:"is_active"`

	Rooms []Room `gorm:"many2many:room_amenities;" json:"rooms,omitempty"`
}

func (Amenity) TableName() string {
	return "amenities"
}
