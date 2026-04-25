package dto

type CreateReservationRequest struct {
	RoomID       string `json:"room_id" binding:"required,uuid"`
	GuestName    string `json:"guest_name" binding:"required,min=2"`
	GuestPhone   string `json:"guest_phone" binding:"required"`
	CheckInDate  string `json:"check_in_date" binding:"required"`  // YYYY-MM-DD
	CheckOutDate string `json:"check_out_date" binding:"required"` // YYYY-MM-DD
	Notes        string `json:"notes"`
}

type ListReservationsQuery struct {
	Status   string `form:"status"`
	RoomID   string `form:"room_id"`
	DateFrom string `form:"date_from"` // YYYY-MM-DD
	DateTo   string `form:"date_to"`   // YYYY-MM-DD
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
}

type AvailabilityQuery struct {
	CheckInDate  string `form:"check_in" binding:"required"`  // YYYY-MM-DD
	CheckOutDate string `form:"check_out" binding:"required"` // YYYY-MM-DD
}
