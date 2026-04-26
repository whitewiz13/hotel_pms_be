package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type GuestSettingsHandler struct {
	settingsService *service.GuestSettingsService
}

func NewGuestSettingsHandler(settingsService *service.GuestSettingsService) *GuestSettingsHandler {
	return &GuestSettingsHandler{settingsService: settingsService}
}

// Get returns guest settings for a hotel (used by both admin and guest portal).
func (h *GuestSettingsHandler) Get(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	settings, err := h.settingsService.Get(hotelID)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch guest settings")
		return
	}

	utils.RespondOK(c, settings)
}

// Save creates or updates guest settings for a hotel (admin only).
func (h *GuestSettingsHandler) Save(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.SaveGuestSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	settings, err := h.settingsService.Save(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, settings)
}

// GetForGuest returns guest settings from the guest's JWT hotel context.
func (h *GuestSettingsHandler) GetForGuest(c *gin.Context) {
	hotelID, _, ok := getGuestContext(c)
	if !ok {
		return
	}

	settings, err := h.settingsService.Get(hotelID)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch guest settings")
		return
	}

	utils.RespondOK(c, settings)
}
