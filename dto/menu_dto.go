package dto

import "github.com/hotelpms/backend/models"

type CreateMenuItemRequest struct {
	Name        string               `json:"name" binding:"required,min=2"`
	Description string               `json:"description"`
	Price       float64              `json:"price" binding:"required,gt=0"`
	Category    models.MenuCategory  `json:"category" binding:"required,oneof=appetizer main_course dessert beverage snack"`
}

type UpdateMenuItemRequest struct {
	Name        *string              `json:"name" binding:"omitempty,min=2"`
	Description *string              `json:"description"`
	Price       *float64             `json:"price" binding:"omitempty,gt=0"`
	Category    *models.MenuCategory `json:"category" binding:"omitempty,oneof=appetizer main_course dessert beverage snack"`
	IsAvailable *bool                `json:"is_available"`
}

type ListMenuQuery struct {
	Category string `form:"category"`
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
}
