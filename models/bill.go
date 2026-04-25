package models

type BillStatus string

const (
	BillStatusPending BillStatus = "pending"
	BillStatusPaid    BillStatus = "paid"
)

type BillLineType string

const (
	BillLineRoom        BillLineType = "room_charge"
	BillLineRoomService BillLineType = "room_service"
	BillLineActivity    BillLineType = "activity"
	BillLineUpfront     BillLineType = "upfront_payment"
)

type Bill struct {
	BaseModel
	HotelID          string     `gorm:"type:uuid;not null;index" json:"hotel_id"`
	ReservationID    string     `gorm:"type:uuid;not null;uniqueIndex" json:"reservation_id"`
	RoomID           string     `gorm:"type:uuid;not null" json:"room_id"`
	GuestName        string     `gorm:"not null;size:255" json:"guest_name"`
	RoomCharges      float64    `gorm:"not null;default:0" json:"room_charges"`
	UpfrontPaid      float64    `gorm:"not null;default:0" json:"upfront_paid"`
	RoomServiceTotal float64    `gorm:"not null;default:0" json:"room_service_total"`
	ActivityTotal    float64    `gorm:"not null;default:0" json:"activity_total"`
	Subtotal         float64    `gorm:"not null;default:0" json:"subtotal"`
	TaxRate          float64    `gorm:"not null;default:0" json:"tax_rate"`
	TaxAmount        float64    `gorm:"not null;default:0" json:"tax_amount"`
	TotalAmount      float64    `gorm:"not null;default:0" json:"total_amount"`
	Status           BillStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Notes            string     `gorm:"type:text" json:"notes"`

	Hotel       Hotel          `gorm:"foreignKey:HotelID" json:"-"`
	Reservation Reservation    `gorm:"foreignKey:ReservationID" json:"reservation,omitempty"`
	Room        Room           `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	LineItems   []BillLineItem `gorm:"foreignKey:BillID" json:"line_items,omitempty"`
}

func (Bill) TableName() string {
	return "bills"
}

type BillLineItem struct {
	BaseModel
	BillID      string       `gorm:"type:uuid;not null;index" json:"bill_id"`
	Type        BillLineType `gorm:"type:varchar(30);not null" json:"type"`
	Description string       `gorm:"not null;size:500" json:"description"`
	Amount      float64      `gorm:"not null;default:0" json:"amount"`
	ReferenceID *string      `gorm:"type:uuid" json:"reference_id,omitempty"`

	Bill Bill `gorm:"foreignKey:BillID" json:"-"`
}

func (BillLineItem) TableName() string {
	return "bill_line_items"
}
