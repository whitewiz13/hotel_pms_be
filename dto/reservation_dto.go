package dto

type CreateReservationRequest struct {
	RoomID       string `json:"room_id" binding:"required,uuid"`
	GuestName    string `json:"guest_name" binding:"required,min=2"`
	GuestPhone   string `json:"guest_phone" binding:"required"`
	GuestEmail   string `json:"guest_email"`
	CheckInDate  string `json:"check_in_date" binding:"required"`  // YYYY-MM-DD
	CheckOutDate string `json:"check_out_date" binding:"required"` // YYYY-MM-DD
	Notes        string `json:"notes"`
}

type ListReservationsQuery struct {
	Status       string `form:"status"`
	RoomID       string `form:"room_id"`
	DateFrom     string `form:"date_from"`      // YYYY-MM-DD range filter: check_in_date >= date_from
	DateTo       string `form:"date_to"`        // YYYY-MM-DD range filter: check_out_date <= date_to
	CheckInDate  string `form:"check_in_date"`  // YYYY-MM-DD exact match on check_in_date
	CheckOutDate string `form:"check_out_date"` // YYYY-MM-DD exact match on check_out_date
	Page         int    `form:"page"`
	PerPage      int    `form:"per_page"`
}

type CheckInRequest struct {
	GuestEmail    *string `json:"guest_email" binding:"omitempty,email"`
	IDType        *string `json:"id_type" binding:"omitempty,oneof=aadhaar pan passport driving_license voter_id"`
	IDNumber      *string `json:"id_number"`
	IDDocumentURL *string `json:"id_document_url"`
}

type AvailabilityQuery struct {
	CheckInDate  string `form:"check_in" binding:"required"`  // YYYY-MM-DD
	CheckOutDate string `form:"check_out" binding:"required"` // YYYY-MM-DD
}
