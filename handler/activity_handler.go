package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// --- Activity CRUD ---

func (h *ActivityHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	activity, err := h.activityService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, activity)
}

func (h *ActivityHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	activity, err := h.activityService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, activity)
}

func (h *ActivityHandler) List(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListActivitiesQuery{
		Category: c.Query("category"),
		Page:     page,
		PerPage:  perPage,
	}

	activities, total, err := h.activityService.List(hotelID, query)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch activities")
		return
	}

	utils.RespondPaginated(c, activities, query.Page, query.PerPage, total)
}

func (h *ActivityHandler) Update(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	activity, err := h.activityService.Update(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, activity)
}

func (h *ActivityHandler) Delete(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	if err := h.activityService.Delete(id, hotelID); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "activity deleted successfully")
}

// --- Activity Booking ---

func (h *ActivityHandler) CreateBooking(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateActivityBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	booking, err := h.activityService.CreateBooking(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, booking)
}

func (h *ActivityHandler) GetBookingByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	booking, err := h.activityService.GetBookingByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, booking)
}

func (h *ActivityHandler) ListBookings(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListActivityBookingsQuery{
		Status:        c.Query("status"),
		RoomID:        c.Query("room_id"),
		ReservationID: c.Query("reservation_id"),
		ActivityID:    c.Query("activity_id"),
		Page:          page,
		PerPage:       perPage,
	}

	bookings, total, err := h.activityService.ListBookings(hotelID, query)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, bookings, query.Page, query.PerPage, total)
}

func (h *ActivityHandler) UpdateBookingStatus(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateActivityBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	booking, err := h.activityService.UpdateBookingStatus(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, booking)
}
