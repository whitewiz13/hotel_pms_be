package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	token, user, err := h.authService.Login(req)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondOK(c, dto.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) GuestLogin(c *gin.Context) {
	var req dto.GuestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	token, room, err := h.authService.GuestLogin(req)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"token":       token,
		"room_number": room.RoomNumber,
		"room_type":   room.RoomType,
	})
}

func (h *AuthHandler) CreateStaff(c *gin.Context) {
	var req dto.CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	user, err := h.authService.CreateStaff(req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, user)
}
