package models

import "time"

// Subscription status constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusCancelled = "cancelled"
)

// Subscription links a hotel to a plan.
type Subscription struct {
	BaseModel
	HotelID   string     `gorm:"uniqueIndex;not null" json:"hotel_id"`
	PlanID    string     `gorm:"not null;size:20" json:"plan_id"`
	Plan      Plan       `gorm:"foreignKey:PlanID" json:"plan"`
	Status    string     `gorm:"not null;size:20;default:'active'" json:"status"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
