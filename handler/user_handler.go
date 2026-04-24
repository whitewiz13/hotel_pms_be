package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	role := c.Query("role")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	users, total, err := h.userService.GetAll(page, perPage, role)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch users")
		return
	}

	utils.RespondPaginated(c, users, page, perPage, total)
}

func (h *UserHandler) GetByHotelID(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	users, total, err := h.userService.GetByHotelID(hotelID, page, perPage)
	if err != nil {
		utils.RespondInternalError(c, "failed to fetch staff")
		return
	}

	utils.RespondPaginated(c, users, page, perPage, total)
}
