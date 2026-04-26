package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type RoleHandler struct {
	roleService *service.RoleService
}

func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	role, err := h.roleService.CreateRole(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, role)
}

func (h *RoleHandler) GetAll(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	roles, err := h.roleService.GetRolesByHotel(hotelID)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch roles")
		return
	}

	utils.RespondOK(c, roles)
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	role, err := h.roleService.GetRoleByID(id)
	if err != nil {
		utils.RespondNotFound(c, "role not found")
		return
	}

	utils.RespondOK(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	role, err := h.roleService.UpdateRole(id, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.roleService.DeleteRole(id); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "role deleted")
}

func (h *RoleHandler) GetPermissions(c *gin.Context) {
	permissions, err := h.roleService.GetAllPermissions()
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch permissions")
		return
	}

	utils.RespondOK(c, permissions)
}
