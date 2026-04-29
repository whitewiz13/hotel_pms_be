package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type ReservationHandler struct {
	reservationService *service.ReservationService
}

func NewReservationHandler(reservationService *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservationService: reservationService}
}

func (h *ReservationHandler) GetAvailability(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var query dto.AvailabilityQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	rooms, err := h.reservationService.GetAvailability(hotelID, query)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, rooms)
}

func (h *ReservationHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	reservation, err := h.reservationService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, reservation)
}

func (h *ReservationHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	reservation, err := h.reservationService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, reservation)
}

func (h *ReservationHandler) List(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListReservationsQuery{
		Status:       c.Query("status"),
		RoomID:       c.Query("room_id"),
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
		CheckInDate:  c.Query("check_in_date"),
		CheckOutDate: c.Query("check_out_date"),
		Page:         page,
		PerPage:      perPage,
	}

	reservations, total, err := h.reservationService.List(hotelID, query)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, reservations, query.Page, query.PerPage, total)
}

func (h *ReservationHandler) CheckIn(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.CheckInRequest
	// Body is optional — allow empty body for backward compatibility
	_ = c.ShouldBindJSON(&req)

	reservation, pin, err := h.reservationService.CheckIn(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"reservation": reservation,
		"access_pin":  pin,
	})
}

func (h *ReservationHandler) CheckOut(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	reservation, err := h.reservationService.CheckOut(id, hotelID)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, reservation)
}

func (h *ReservationHandler) Cancel(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	reservation, err := h.reservationService.Cancel(id, hotelID)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, reservation)
}
