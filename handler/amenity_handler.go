package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type AmenityHandler struct {
	amenityService *service.AmenityService
}

func NewAmenityHandler(amenityService *service.AmenityService) *AmenityHandler {
	return &AmenityHandler{amenityService: amenityService}
}

func (h *AmenityHandler) Create(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	var req dto.CreateAmenityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	amenity, err := h.amenityService.Create(hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, amenity)
}

func (h *AmenityHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	amenity, err := h.amenityService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, amenity)
}

func (h *AmenityHandler) GetAll(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	category := c.Query("category")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var amenities []models.Amenity
	var total int64
	var err error

	if category != "" {
		amenities, total, err = h.amenityService.GetByCategoryAndHotel(category, hotelID, page, perPage)
	} else {
		amenities, total, err = h.amenityService.GetByHotelID(hotelID, page, perPage)
	}

	if err != nil {
		utils.RespondInternalError(c, "failed to fetch amenities")
		return
	}

	utils.RespondPaginated(c, amenities, page, perPage, total)
}

func (h *AmenityHandler) Update(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	var req dto.UpdateAmenityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	amenity, err := h.amenityService.Update(id, hotelID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, amenity)
}

func (h *AmenityHandler) Delete(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	if err := h.amenityService.Delete(id, hotelID); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondMessage(c, "amenity deleted successfully")
}
