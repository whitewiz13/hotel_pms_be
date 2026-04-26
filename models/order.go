package models

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	BaseModel
	HotelID       string      `gorm:"type:uuid;not null;index" json:"hotel_id"`
	RoomID        string      `gorm:"type:uuid;not null;index" json:"room_id"`
	ReservationID string      `gorm:"type:uuid;not null;index" json:"reservation_id"`
	GuestID       string      `gorm:"type:uuid;not null;index" json:"guest_id"`
	GuestName     string      `gorm:"not null;size:255" json:"guest_name"`
	Status        OrderStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	TotalAmount   float64     `gorm:"not null;default:0" json:"total_amount"`
	Notes         string      `gorm:"type:text" json:"notes"`
	AssignedToID  *string     `gorm:"type:uuid;index" json:"assigned_to_id,omitempty"`

	Hotel       Hotel       `gorm:"foreignKey:HotelID" json:"-"`
	Room        Room        `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Reservation Reservation `gorm:"foreignKey:ReservationID" json:"-"`
	Guest       Guest       `gorm:"foreignKey:GuestID" json:"guest,omitempty"`
	AssignedTo  *User       `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	BaseModel
	OrderID    string  `gorm:"type:uuid;not null;index" json:"order_id"`
	MenuItemID string  `gorm:"type:uuid;not null" json:"menu_item_id"`
	Quantity   int     `gorm:"not null;default:1" json:"quantity"`
	UnitPrice  float64 `gorm:"not null" json:"unit_price"`
	Subtotal   float64 `gorm:"not null" json:"subtotal"`
	Notes      string  `gorm:"type:text" json:"notes"`

	Order    Order    `gorm:"foreignKey:OrderID" json:"-"`
	MenuItem MenuItem `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
