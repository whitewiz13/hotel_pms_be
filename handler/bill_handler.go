package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type BillHandler struct {
	billService *service.BillService
}

func NewBillHandler(billService *service.BillService) *BillHandler {
	return &BillHandler{billService: billService}
}

func (h *BillHandler) Generate(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	reservationID := c.Param("id")

	var req dto.GenerateBillRequest
	_ = c.ShouldBindJSON(&req) // Optional body

	bill, err := h.billService.Generate(hotelID, reservationID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, bill)
}

func (h *BillHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	bill, err := h.billService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, bill)
}

func (h *BillHandler) GetByReservation(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	reservationID := c.Param("id")

	bill, err := h.billService.GetByReservationID(reservationID, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, bill)
}

func (h *BillHandler) List(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListBillsQuery{
		Status:        c.Query("status"),
		ReservationID: c.Query("reservation_id"),
		Page:          page,
		PerPage:       perPage,
	}

	bills, total, err := h.billService.List(hotelID, query)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, bills, query.Page, query.PerPage, total)
}

func (h *BillHandler) MarkPaid(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	bill, err := h.billService.MarkPaid(id, hotelID)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, bill)
}
