package handler

import (
	"fmt"
	"log"
	"time"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/service"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

type AspirationHandler struct {
	service    service.AspirationService
	sseManager *SSEManager
}

func NewAspirationHandler(service service.AspirationService, sseManager *SSEManager) *AspirationHandler {
	return &AspirationHandler{service, sseManager}
}

// Create a New Aspiration
func (h *AspirationHandler) Create(c *gin.Context) {
	var request dto.AspirationRequest
	err := c.ShouldBind(&request)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request body")); return }

	userID := c.GetUint("userID")
	ctx := c.Request.Context()
	aspiration, err := h.service.Create(ctx, userID, &request)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Aspiration created successfully",
		Data:    aspiration,
	})
}

// Delete a Aspiration
func (h *AspirationHandler) Delete(c *gin.Context) {
	param := c.Param("id")
	aspirationID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	ctx := c.Request.Context()
	err = h.service.Delete(ctx, aspirationID)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Aspiration deleted successfully",
	})
}

// Get All Aspirations
func (h *AspirationHandler) GetAll(c *gin.Context) {
	var filter dto.AspirationFilterRequest
	err := c.ShouldBindQuery(&filter)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	sessionID := c.GetUint("userID")
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }
	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	aspirations, pagination, err := h.service.GetAll(ctx, sessionID, role, &filter)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success:    true,
		Message:    "Aspirations retrieved successfully",
		Data:       aspirations,
		Pagination: pagination,
	})
}

// Get Aspiration by ID
func (h *AspirationHandler) GetByID(c *gin.Context) {
	param := c.Param("id")
	aspirationID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }
	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	aspiration, err := h.service.GetByID(ctx, aspirationID, sessionID, role)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Aspiration retrieved successfully",
		Data:    aspiration,
	})
}

// Get All Aspirations by Session User
func (h *AspirationHandler) GetAllBySession(c *gin.Context) {
	var filter dto.AspirationFilterRequest
	err := c.ShouldBindQuery(&filter)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	userID := c.GetUint("userID")
	filter.SenderID = &userID
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }
	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	aspirations, pagination, err := h.service.GetAll(ctx, userID, role, &filter)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success:    true,
		Message:    "Aspirations retrieved successfully",
		Data:       aspirations,
		Pagination: pagination,
	})
}

// Export Aspirations to Excel
func (h *AspirationHandler) Export(c *gin.Context) {
	var filter dto.AspirationFilterRequest
	err := c.ShouldBindQuery(&filter)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	ctx := c.Request.Context()
	excel, err := h.service.Export(ctx, &filter)
	if err != nil { c.Error(err); return }

	now := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("Export Aspirations_%s.xlsx", now)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excel.Bytes())
}

// Close an Aspiration
func (h *AspirationHandler) Close(c *gin.Context) {
	param := c.Param("id")
	aspirationID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }
	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	err = h.service.Close(ctx, aspirationID, sessionID, role)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Aspiration closed successfully",
	})
}

// Create a New Vote
func (h *AspirationHandler) Vote(c *gin.Context) {
	var request dto.VoteRequest
	err := c.ShouldBindJSON(&request)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request body")); return }

	param := c.Param("id")
	aspirationID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }
	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	vote, err := h.service.Vote(ctx, aspirationID, sessionID, role, &request)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Vote created successfully",
		Data:    vote,
	})
}

// Create a New Discussion
func (h *AspirationHandler) Discussion(c *gin.Context) {
	var request dto.DiscussionRequest
	err := c.ShouldBind(&request)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request body")); return }

	param := c.Param("id")
	aspirationID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }
	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	discussionResponse, err := h.service.Discussion(ctx, aspirationID, sessionID, role, &request)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Discussion created successfully",
		Data:    discussionResponse,
	})
}

// Stream Discussion Real-time Events
func (h *AspirationHandler) Stream(c *gin.Context) {
	param := c.Param("id")
	aspirationID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	sessionRole, exists := c.Get("userRole")
	if !exists { c.Error(apperror.ErrForbidden("Role not found in token")); return }

	role, ok := sessionRole.(model.RoleEnum)
	if !ok { c.Error(apperror.ErrForbidden("Invalid role in token")); return }

	ctx := c.Request.Context()
	_, err = h.service.GetByID(ctx, aspirationID, sessionID, role)
	if err != nil { c.Error(err); return }

	client := h.sseManager.AddClient(aspirationID, sessionID, role)
	log.Printf("[SSE] User %d (Role: %s) connected to aspiration %d stream", sessionID, role, aspirationID)
	defer func() {
		h.sseManager.RemoveClient(aspirationID, client)
		log.Printf("[SSE] User %d (Role: %s) disconnected from aspiration %d stream", sessionID, role, aspirationID)
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	c.SSEvent("connected", map[string]interface{}{
		"message":       "Successfully connected to aspiration stream",
		"aspiration_id": aspirationID,
	})
	
	c.Writer.Flush()
	clientGone := c.Request.Context().Done()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientGone: return
		case <-ticker.C:
			if _, err := c.Writer.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			c.Writer.Flush()
		case msg, ok := <-client.MessageChan:
			if !ok { return }

			c.SSEvent(msg.Event, msg.Data)
			c.Writer.Flush()
		}
	}
}

