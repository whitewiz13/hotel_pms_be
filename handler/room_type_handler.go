package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type RoomTypeHandler struct {
	roomTypeService *service.RoomTypeService
}

func NewRoomTypeHandler(roomTypeService *service.RoomTypeService) *RoomTypeHandler {
	return &RoomTypeHandler{roomTypeService: roomTypeService}
}

func (h *RoomTypeHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateRoomTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	roomType, err := h.roomTypeService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, roomType)
}

func (h *RoomTypeHandler) GetAll(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	roomTypes, err := h.roomTypeService.GetByHotelID(hotelID)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch room types")
		return
	}

	utils.RespondOK(c, roomTypes)
}

func (h *RoomTypeHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	roomType, err := h.roomTypeService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, roomType)
}

func (h *RoomTypeHandler) Update(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateRoomTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	roomType, err := h.roomTypeService.Update(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, roomType)
}

func (h *RoomTypeHandler) Delete(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	if err := h.roomTypeService.Delete(id, hotelID); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "room type deleted successfully")
}
