package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func (h *RoomHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	room, err := h.roomService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, room)
}

func (h *RoomHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	room, err := h.roomService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, room)
}

func (h *RoomHandler) GetByHotelID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	status := c.Query("status")

	rooms, total, err := h.roomService.GetByHotelID(hotelID, status, page, perPage)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch rooms")
		return
	}

	utils.RespondPaginated(c, rooms, page, perPage, total)
}

func (h *RoomHandler) Update(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	room, err := h.roomService.Update(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, room)
}

func (h *RoomHandler) Delete(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	if err := h.roomService.Delete(id, hotelID); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "room deleted successfully")
}

func (h *RoomHandler) GeneratePin(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	pin, err := h.roomService.SetAccessPin(id, hotelID)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, map[string]string{"pin": pin})
}

func (h *RoomHandler) ClearPin(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	if err := h.roomService.ClearAccessPin(id, hotelID); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "room access pin cleared")
}
