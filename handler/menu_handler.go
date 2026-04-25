package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type MenuHandler struct {
	menuService *service.MenuService
}

func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

func (h *MenuHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	item, err := h.menuService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, item)
}

func (h *MenuHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	item, err := h.menuService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, item)
}

func (h *MenuHandler) List(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListMenuQuery{
		Category: c.Query("category"),
		Page:     page,
		PerPage:  perPage,
	}

	items, total, err := h.menuService.List(hotelID, query)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch menu items")
		return
	}

	utils.RespondPaginated(c, items, query.Page, query.PerPage, total)
}

func (h *MenuHandler) Update(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	item, err := h.menuService.Update(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, item)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	if err := h.menuService.Delete(id, hotelID); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "menu item deleted successfully")
}
