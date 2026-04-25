package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	stats, err := h.dashboardService.GetStats(hotelID)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch dashboard stats")
		return
	}

	utils.RespondOK(c, stats)
}

func (h *DashboardHandler) GetActivity(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	activity, err := h.dashboardService.GetActivity(hotelID, limit)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch activity feed")
		return
	}

	utils.RespondOK(c, activity)
}
