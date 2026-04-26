package models

type Hotel struct {
	BaseModel
	Name        string `gorm:"not null;size:255" json:"name"`
	Slug        string `gorm:"uniqueIndex;not null;size:100" json:"slug"`
	Address     string `gorm:"not null;size:500" json:"address"`
	City        string `gorm:"not null;size:100" json:"city"`
	State       string `gorm:"size:100" json:"state"`
	Country     string `gorm:"not null;size:100" json:"country"`
	ZipCode     string `gorm:"size:20" json:"zip_code"`
	Phone       string `gorm:"size:20" json:"phone"`
	Email       string `gorm:"size:255" json:"email"`
	Description string `gorm:"type:text" json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	Rooms     []Room    `gorm:"foreignKey:HotelID" json:"rooms,omitempty"`
	Staff     []User    `gorm:"foreignKey:HotelID" json:"staff,omitempty"`
	Amenities []Amenity `gorm:"foreignKey:HotelID" json:"amenities,omitempty"`
}

func (Hotel) TableName() string {
	return "hotels"
}
