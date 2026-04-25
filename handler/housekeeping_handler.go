package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/middleware"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type HousekeepingHandler struct {
	housekeepingService *service.HousekeepingService
}

func NewHousekeepingHandler(housekeepingService *service.HousekeepingService) *HousekeepingHandler {
	return &HousekeepingHandler{housekeepingService: housekeepingService}
}

// Assign creates a new housekeeping task and sets the room to "cleaning".
func (h *HousekeepingHandler) Assign(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	claims := middleware.GetClaims(c)

	var req dto.AssignHousekeepingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	task, err := h.housekeepingService.Assign(hotelID, claims.UserID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondCreated(c, task)
}

// Complete marks a task as done and sets the room to "available".
func (h *HousekeepingHandler) Complete(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")
	claims := middleware.GetClaims(c)

	var req dto.UpdateHousekeepingTaskRequest
	// Body is optional for completion
	_ = c.ShouldBindJSON(&req)

	task, err := h.housekeepingService.Complete(id, hotelID, claims.UserID, req)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, task)
}

// GetByID returns a single housekeeping task.
func (h *HousekeepingHandler) GetByID(c *gin.Context) {
	hotelID := c.Param("hotel_id")
	id := c.Param("id")

	task, err := h.housekeepingService.GetByID(id, hotelID)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, task)
}

// List returns housekeeping tasks ordered by priority.
func (h *HousekeepingHandler) List(c *gin.Context) {
	hotelID := c.Param("hotel_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	query := dto.ListHousekeepingQuery{
		Status:       c.Query("status"),
		AssignedToID: c.Query("assigned_to_id"),
		RoomID:       c.Query("room_id"),
		Priority:     c.Query("priority"),
		Page:         page,
		PerPage:      perPage,
	}

	tasks, total, err := h.housekeepingService.List(hotelID, query)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondPaginated(c, tasks, query.Page, query.PerPage, total)
}
