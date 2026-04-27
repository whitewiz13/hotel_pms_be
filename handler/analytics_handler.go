package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// parseDateRange extracts "from" and "to" query params, defaulting to last 30 days.
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	from := now.AddDate(0, -1, 0).Truncate(24 * time.Hour)
	to := now.Truncate(24*time.Hour).Add(24*time.Hour - time.Second)

	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(24*time.Hour - time.Second)
		}
	}
	return from, to
}

// parsePeriod reads the "period" query param (daily, monthly, yearly). Defaults to daily.
func parsePeriod(c *gin.Context) string {
	p := c.DefaultQuery("period", "daily")
	switch p {
	case "daily", "monthly", "yearly":
		return p
	default:
		return "daily"
	}
}

func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	from, to := parseDateRange(c)

	summary, err := h.analyticsService.GetSummary(hotelID, from, to)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch analytics summary")
		return
	}

	utils.RespondOK(c, summary)
}

func (h *AnalyticsHandler) GetOccupancyTrend(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	from, to := parseDateRange(c)
	period := parsePeriod(c)

	data, err := h.analyticsService.GetOccupancyTrend(hotelID, from, to, period)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch occupancy trend")
		return
	}

	utils.RespondOK(c, data)
}

func (h *AnalyticsHandler) GetRevenueTrend(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	from, to := parseDateRange(c)
	period := parsePeriod(c)

	data, err := h.analyticsService.GetRevenueTrend(hotelID, from, to, period)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch revenue trend")
		return
	}

	utils.RespondOK(c, data)
}

func (h *AnalyticsHandler) GetReservationStats(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	from, to := parseDateRange(c)
	period := parsePeriod(c)

	data, err := h.analyticsService.GetReservationStats(hotelID, from, to, period)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch reservation stats")
		return
	}

	utils.RespondOK(c, data)
}

func (h *AnalyticsHandler) GetRoomTypePerformance(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	from, to := parseDateRange(c)

	data, err := h.analyticsService.GetRoomTypePerformance(hotelID, from, to)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch room type performance")
		return
	}

	utils.RespondOK(c, data)
}
