package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/middleware"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
)

type FCMTokenHandler struct {
	fcmTokenRepo *repository.FCMTokenRepository
}

func NewFCMTokenHandler(fcmTokenRepo *repository.FCMTokenRepository) *FCMTokenHandler {
	return &FCMTokenHandler{fcmTokenRepo: fcmTokenRepo}
}

// SaveToken stores or refreshes an FCM device token for the authenticated user.
func (h *FCMTokenHandler) SaveToken(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		utils.RespondUnauthorized(c, "authentication required")
		return
	}

	var req dto.SaveFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	token := &models.FCMToken{
		UserID:      claims.UserID,
		HotelID:     claims.HotelID,
		DeviceToken: req.DeviceToken,
	}

	if err := h.fcmTokenRepo.Upsert(token); err != nil {
		utils.RespondInternalError(c, "failed to save device token")
		return
	}

	utils.RespondOK(c, gin.H{"message": "device token saved"})
}

// DeleteToken removes a specific device token (e.g. on logout).
func (h *FCMTokenHandler) DeleteToken(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		utils.RespondUnauthorized(c, "authentication required")
		return
	}

	var req dto.SaveFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.fcmTokenRepo.DeleteByToken(claims.UserID, req.DeviceToken); err != nil {
		utils.RespondInternalError(c, "failed to remove device token")
		return
	}

	utils.RespondMessage(c, "device token removed")
}
