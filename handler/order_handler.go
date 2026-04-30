package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/middleware"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type OrderHandler struct {
	orderService *service.OrderService
	roleRepo     *repository.RoleRepository
}

func NewOrderHandler(orderService *service.OrderService, roleRepo *repository.RoleRepository) *OrderHandler {
	return &OrderHandler{orderService: orderService, roleRepo: roleRepo}
}

func (h *OrderHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.orderService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	order, err := h.orderService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	claims := middleware.GetClaims(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListOrdersQuery{
		Status:        c.Query("status"),
		RoomID:        c.Query("room_id"),
		ReservationID: c.Query("reservation_id"),
		AssignedToID:  c.Query("assigned_to_id"),
		Page:          page,
		PerPage:       perPage,
	}

	// Users without orders:assign only see orders assigned to them
	if !middleware.HasPermission(c, h.roleRepo, "orders:assign") {
		query.AssignedToID = claims.UserID
	}

	orders, total, err := h.orderService.List(hotelID, query)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, orders, query.Page, query.PerPage, total)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.orderService.UpdateStatus(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, order)
}

func (h *OrderHandler) Assign(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.AssignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.orderService.Assign(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, order)
}
