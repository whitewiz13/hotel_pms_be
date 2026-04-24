package dto

import "github.com/hotelpms/backend/models"

type CreateAmenityRequest struct {
	Name        string                 `json:"name" binding:"required,min=2"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Category    models.AmenityCategory `json:"category" binding:"required,oneof=room bathroom general dining recreation"`
}

type UpdateAmenityRequest struct {
	Name        *string                 `json:"name" binding:"omitempty,min=2"`
	Description *string                 `json:"description"`
	Icon        *string                 `json:"icon"`
	Category    *models.AmenityCategory `json:"category" binding:"omitempty,oneof=room bathroom general dining recreation"`
	IsActive    *bool                   `json:"is_active"`
}
