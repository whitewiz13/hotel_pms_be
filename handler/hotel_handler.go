package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type HotelHandler struct {
	hotelService *service.HotelService
}

func NewHotelHandler(hotelService *service.HotelService) *HotelHandler {
	return &HotelHandler{hotelService: hotelService}
}

func (h *HotelHandler) Create(c *gin.Context) {
	var req dto.CreateHotelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	hotel, admin, err := h.hotelService.Create(req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, gin.H{
		"hotel": hotel,
		"admin": admin,
	})
}

func (h *HotelHandler) GetByID(c *gin.Context) {
	id := c.Param("hotel_id")

	hotel, err := h.hotelService.GetByID(id)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, hotel)
}

func (h *HotelHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	hotels, total, err := h.hotelService.GetAll(page, perPage)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch hotels")
		return
	}

	utils.RespondPaginated(c, hotels, page, perPage, total)
}

func (h *HotelHandler) Update(c *gin.Context) {
	id := c.Param("hotel_id")

	var req dto.UpdateHotelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	hotel, err := h.hotelService.Update(id, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, hotel)
}

func (h *HotelHandler) Delete(c *gin.Context) {
	id := c.Param("hotel_id")

	if err := h.hotelService.Delete(id); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "hotel deleted successfully")
}
