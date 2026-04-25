package dto

import "github.com/hotelpms/backend/models"

type AssignHousekeepingRequest struct {
	RoomID       string                       `json:"room_id" binding:"required,uuid"`
	AssignedToID *string                      `json:"assigned_to_id" binding:"omitempty,uuid"`
	Priority     models.HousekeepingPriority  `json:"priority" binding:"required,oneof=low normal high urgent"`
	Notes        string                       `json:"notes"`
}

type UpdateHousekeepingTaskRequest struct {
	Notes string `json:"notes"`
}

type ListHousekeepingQuery struct {
	Status       string `form:"status"`
	AssignedToID string `form:"assigned_to_id"`
	RoomID       string `form:"room_id"`
	Priority     string `form:"priority"`
	Page         int    `form:"page"`
	PerPage      int    `form:"per_page"`
}
