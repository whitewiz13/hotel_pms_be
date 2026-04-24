package dto

type CreateHotelRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Address     string `json:"address" binding:"required"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state"`
	Country     string `json:"country" binding:"required"`
	ZipCode     string `json:"zip_code"`
	Phone       string `json:"phone"`
	Email       string `json:"email" binding:"omitempty,email"`
	Description string `json:"description"`

	// Hotel admin credentials
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=8"`
	AdminName     string `json:"admin_name" binding:"required,min=2"`
}

type UpdateHotelRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Address     *string `json:"address"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	Country     *string `json:"country"`
	ZipCode     *string `json:"zip_code"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email" binding:"omitempty,email"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}
