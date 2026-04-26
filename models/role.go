package models

// Role represents a dynamic role that can be assigned to users.
type Role struct {
	BaseModel
	HotelID     *string      `gorm:"type:uuid;index" json:"hotel_id,omitempty"`
	Name        string       `gorm:"not null;size:100" json:"name"`
	Slug        string       `gorm:"not null;size:100" json:"slug"`
	Description string       `gorm:"size:255" json:"description"`
	IsSystem    bool         `gorm:"default:false" json:"is_system"`
	Hotel       *Hotel       `gorm:"foreignKey:HotelID" json:"hotel,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}


