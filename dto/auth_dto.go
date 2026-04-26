package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type GuestLoginRequest struct {
	RoomNumber string `json:"room_number" binding:"required"`
	Pin        string `json:"pin" binding:"required,len=6"`
	HotelID    string `json:"hotel_id" binding:"required,uuid"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

type CreateStaffRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
	Phone    string `json:"phone"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
}
