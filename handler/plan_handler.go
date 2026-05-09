package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type PlanHandler struct {
	planService *service.PlanService
}

func NewPlanHandler(planService *service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

// GetAllPlans returns all available plans.
func (h *PlanHandler) GetAllPlans(c *gin.Context) {
	plans, err := h.planService.GetAllPlans()
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch plans")
		return
	}
	utils.RespondOK(c, plans)
}

// GetHotelSubscription returns the current subscription state for a hotel.
func (h *PlanHandler) GetHotelSubscription(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	sub, err := h.planService.GetHotelSubscriptionDetails(hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}
	utils.RespondOK(c, sub)
}

// ChangeHotelPlan updates the subscription state for a hotel (super_admin only).
func (h *PlanHandler) ChangeHotelPlan(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.UpdateHotelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	sub, err := h.planService.UpdateHotelSubscription(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, sub)
}

// GetHotelPlanUsage returns current usage vs limits for a hotel.
func (h *PlanHandler) GetHotelPlanUsage(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	usage, err := h.planService.GetUsage(hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, usage)
}

