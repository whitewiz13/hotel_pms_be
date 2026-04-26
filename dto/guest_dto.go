package dto

type GuestPlaceOrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	Notes string             `json:"notes"`
}

type GuestBookActivityRequest struct {
	ActivityID  string `json:"activity_id" binding:"required,uuid"`
	ScheduledAt string `json:"scheduled_at"`
	Notes       string `json:"notes"`
}
