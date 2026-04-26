package dto

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=100"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

type UpdateRoleRequest struct {
	Name        *string  `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required,uuid"`
}
