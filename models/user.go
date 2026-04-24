package models

type UserRole string

const (
	RoleSuperAdmin  UserRole = "super_admin"
	RoleManager     UserRole = "manager"
	RoleFrontDesk   UserRole = "front_desk"
	RoleHousekeeping UserRole = "housekeeping"
	RoleStaff       UserRole = "staff"
	RoleGuest       UserRole = "guest" // virtual role for JWT only, not stored
)

// ValidRoles lists every role that can be assigned to a user.
// Add or remove a line here to support a new role everywhere.
var ValidRoles = []UserRole{
	RoleSuperAdmin,
	RoleManager,
	RoleFrontDesk,
	RoleHousekeeping,
	RoleStaff,
}

func IsValidRole(r UserRole) bool {
	for _, v := range ValidRoles {
		if v == r {
			return true
		}
	}
	return false
}

type User struct {
	BaseModel
	Email        string   `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string   `gorm:"not null" json:"-"`
	Name         string   `gorm:"not null;size:255" json:"name"`
	Phone        string   `gorm:"size:20" json:"phone"`
	Role         UserRole `gorm:"type:varchar(30);not null;default:'staff'" json:"role"`
	HotelID      *string  `gorm:"type:uuid;index" json:"hotel_id,omitempty"`
	IsActive     bool     `gorm:"default:true" json:"is_active"`

	Hotel *Hotel `gorm:"foreignKey:HotelID" json:"hotel,omitempty"`
}

func (User) TableName() string {
	return "users"
}
