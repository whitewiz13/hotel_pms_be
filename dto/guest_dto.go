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

type GuestDashboardResponse struct {
	RoomNumber       string         `json:"room_number"`
	GuestName        string         `json:"guest_name"`
	CheckInDate      string         `json:"check_in_date"`
	CheckOutDate     string         `json:"check_out_date"`
	OrderStats       map[string]int64 `json:"order_stats"`
	TotalOrders      int64          `json:"total_orders"`
	OrderSpend       float64        `json:"order_spend"`
	ActivityStats    map[string]int64 `json:"activity_stats"`
	TotalActivities  int64          `json:"total_activities"`
	ActivitySpend    float64        `json:"activity_spend"`
}
