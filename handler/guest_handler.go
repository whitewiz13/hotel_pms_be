package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/middleware"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type GuestHandler struct {
	guestService *service.GuestService
}

func NewGuestHandler(guestService *service.GuestService) *GuestHandler {
	return &GuestHandler{guestService: guestService}
}

// getGuestContext extracts hotel_id and room_id from the guest JWT claims.
func getGuestContext(c *gin.Context) (hotelID, roomID string, ok bool) {
	claims := middleware.GetClaims(c)
	if claims == nil || !claims.IsGuest {
		utils.RespondUnauthorized(c, "guest authentication required")
		return "", "", false
	}
	return claims.HotelID, claims.RoomID, true
}

func (h *GuestHandler) GetMyReservation(c *gin.Context) {
	hotelID, roomID, ok := getGuestContext(c)
	if !ok {
		return
	}

	reservation, err := h.guestService.GetMyReservation(roomID, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, reservation)
}

func (h *GuestHandler) ListMenu(c *gin.Context) {
	hotelID, _, ok := getGuestContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	items, total, err := h.guestService.ListMenu(hotelID, page, perPage)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, items, page, perPage, total)
}

func (h *GuestHandler) ListActivities(c *gin.Context) {
	hotelID, _, ok := getGuestContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	activities, total, err := h.guestService.ListActivities(hotelID, page, perPage)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, activities, page, perPage, total)
}

func (h *GuestHandler) PlaceOrder(c *gin.Context) {
	hotelID, roomID, ok := getGuestContext(c)
	if !ok {
		return
	}

	var req dto.GuestPlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.guestService.PlaceOrder(hotelID, roomID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, order)
}

func (h *GuestHandler) ListMyOrders(c *gin.Context) {
	hotelID, roomID, ok := getGuestContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	orders, total, err := h.guestService.ListMyOrders(hotelID, roomID, page, perPage)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, orders, page, perPage, total)
}

func (h *GuestHandler) BookActivity(c *gin.Context) {
	hotelID, roomID, ok := getGuestContext(c)
	if !ok {
		return
	}

	var req dto.GuestBookActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	booking, err := h.guestService.BookActivity(hotelID, roomID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, booking)
}

func (h *GuestHandler) ListMyActivityBookings(c *gin.Context) {
	hotelID, roomID, ok := getGuestContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	bookings, total, err := h.guestService.ListMyActivityBookings(hotelID, roomID, page, perPage)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, bookings, page, perPage, total)
}

func (h *GuestHandler) GetDashboard(c *gin.Context) {
	hotelID, roomID, ok := getGuestContext(c)
	if !ok {
		return
	}

	dashboard, err := h.guestService.GetDashboard(hotelID, roomID)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, dashboard)
}
