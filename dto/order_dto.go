package dto

type OrderItemRequest struct {
	MenuItemID string `json:"menu_item_id" binding:"required,uuid"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	Notes      string `json:"notes"`
}

type CreateOrderRequest struct {
	RoomID        string             `json:"room_id" binding:"required,uuid"`
	ReservationID string             `json:"reservation_id" binding:"required,uuid"`
	GuestName     string             `json:"guest_name" binding:"required,min=2"`
	GuestID       string             `json:"guest_id"`
	Items         []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	Notes         string             `json:"notes"`
	AssignedToID  string             `json:"assigned_to_id" binding:"omitempty,uuid"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=preparing ready delivered cancelled"`
}

type AssignOrderRequest struct {
	AssignedToID string `json:"assigned_to_id" binding:"required,uuid"`
}

type ListOrdersQuery struct {
	Status        string `form:"status"`
	RoomID        string `form:"room_id"`
	ReservationID string `form:"reservation_id"`
	AssignedToID  string `form:"assigned_to_id"`
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
}
