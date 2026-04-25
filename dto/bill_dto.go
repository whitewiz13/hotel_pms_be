package dto

type GenerateBillRequest struct {
	UpfrontPaid float64 `json:"upfront_paid"`
	TaxRate     float64 `json:"tax_rate"`
	Notes       string  `json:"notes"`
}

type UpdateBillStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=paid"`
}

type ListBillsQuery struct {
	Status        string `form:"status"`
	ReservationID string `form:"reservation_id"`
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
}
