package models

import "time"

// Subscription status constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusSuspended = "suspended"
	SubscriptionStatusCancelled = "cancelled"
)

// Subscription links a hotel to a plan.
type Subscription struct {
	BaseModel
	HotelID   string     `gorm:"uniqueIndex;not null" json:"hotel_id"`
	PlanID    string     `gorm:"not null;size:20" json:"plan_id"`
	Plan      Plan       `gorm:"foreignKey:PlanID" json:"plan"`
	Status    string     `gorm:"not null;size:20;default:'active'" json:"status"`
	AssignedAt *time.Time `json:"assigned_at"`
	RenewedAt  *time.Time `json:"renewed_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	SuspendedAt *time.Time `json:"suspended_at"`
	SuspensionReason *string `gorm:"type:text" json:"suspension_reason"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
