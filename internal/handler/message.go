package handler

import (
	"fmt"
	"time"

	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/service"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	service service.MessageService
}

func NewMessageHandler(service service.MessageService) *MessageHandler {
	return &MessageHandler{ service }
}

// Create a New Message
func (h *MessageHandler) Create(c *gin.Context) {
	var request model.MessageRequest
	err := c.ShouldBindJSON(&request);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request body")); return }

	userID := c.GetUint("userID")
	ctx := c.Request.Context()
	message, err := h.service.Create(ctx, userID, &request)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "Message created successfully",
		Data:    message,
	})
}

// Delete a Message
func (h *MessageHandler) Delete(c *gin.Context) {
	param := c.Param("id")
	messageID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	ctx := c.Request.Context()
	err = h.service.Delete(ctx, messageID)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "Message deleted successfully",
	})
}

// Get All Messages
func (h *MessageHandler) GetAll(c *gin.Context) {
	var filter model.MessageFilterRequest
	err := c.ShouldBindQuery(&filter);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	ctx := c.Request.Context()
	messages, pagination, err := h.service.GetAll(ctx, filter)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success:    true,
		Message:    "Messages retrieved successfully",
		Data:       messages,
		Pagination: pagination,
	})
}

// Get Message by ID
func (h *MessageHandler) GetByID(c *gin.Context) {
	param := c.Param("id")
	messageID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	ctx := c.Request.Context()
	message, err := h.service.GetByID(ctx, messageID)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "Message retrieved successfully",
		Data:    message,
	})
}

// Get All Messages by Session User
func (h *MessageHandler) GetAllBySession(c *gin.Context) {
	var filter model.MessageFilterRequest
	err := c.ShouldBindQuery(&filter);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	userID := c.GetUint("userID")
	filter.UserID = &userID

	ctx := c.Request.Context()
	messages, pagination, err := h.service.GetAll(ctx, filter)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "Messages retrieved successfully",
		Data:    messages,
		Pagination: pagination,
	})
}

// Export Messages to Excel
func (h *MessageHandler) Export(c *gin.Context) {
	var filter model.MessageFilterRequest
	err := c.ShouldBindQuery(&filter);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	ctx := c.Request.Context()
	excel, err := h.service.Export(ctx, filter)
	if err != nil { c.Error(err); return }

	now := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("Export Messages_%s.xlsx", now)
    c.Header("Content-Disposition", "attachment; filename=" + filename)
    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excel.Bytes())
}
