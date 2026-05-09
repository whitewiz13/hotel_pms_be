package dto

import (
	"time"

	"github.com/hotelpms/backend/models"
)

type UpdateHotelSubscriptionRequest struct {
	PlanID           *string    `json:"plan_id"`
	Status           *string    `json:"status" binding:"omitempty,oneof=active past_due suspended cancelled"`
	AssignedAt       *time.Time `json:"assigned_at"`
	RenewedAt        *time.Time `json:"renewed_at"`
	AccessUntil      *time.Time `json:"access_until"`
	SuspendedAt      *time.Time `json:"suspended_at"`
	SuspensionReason *string    `json:"suspension_reason"`
	HotelIsActive    *bool      `json:"hotel_is_active"`
}

func (r UpdateHotelSubscriptionRequest) HasChanges() bool {
	return r.PlanID != nil ||
		r.Status != nil ||
		r.AssignedAt != nil ||
		r.RenewedAt != nil ||
		r.AccessUntil != nil ||
		r.SuspendedAt != nil ||
		r.SuspensionReason != nil ||
		r.HotelIsActive != nil
}

type SubscriptionResponse struct {
	ID               string      `json:"id"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	HotelID          string      `json:"hotel_id"`
	PlanID           string      `json:"plan_id"`
	Plan             models.Plan `json:"plan"`
	Status           string      `json:"status"`
	AssignedAt       *time.Time  `json:"assigned_at"`
	RenewedAt        *time.Time  `json:"renewed_at"`
	AccessUntil      *time.Time  `json:"access_until"`
	ExpiresAt        *time.Time  `json:"expires_at"`
	SuspendedAt      *time.Time  `json:"suspended_at"`
	SuspensionReason *string     `json:"suspension_reason"`
	DaysLeft         *int        `json:"days_left"`
	IsOverdue        bool        `json:"is_overdue"`
	OverdueDays      int         `json:"overdue_days"`
	HotelIsActive    bool        `json:"hotel_is_active"`
	CanAccess        bool        `json:"can_access"`
}
