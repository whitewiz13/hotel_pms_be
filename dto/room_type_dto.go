package dto

type CreateRoomTypeRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description"`
}

type UpdateRoomTypeRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=50"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}
