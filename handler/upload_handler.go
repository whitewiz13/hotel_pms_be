package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.RespondBadRequest(c, "file is required")
		return
	}
	defer file.Close()

	folder := c.DefaultPostForm("folder", "uploads")

	url, err := h.uploadService.UploadFile(file, header, folder)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, dto.UploadResponse{URL: url})
}
