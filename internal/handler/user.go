package handler

import (
	"fmt"
	"time"

	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/service"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{ service }
}

// Create a New User
func (h *UserHandler) Create(c *gin.Context) {
	var request model.UserRequest
	err := c.ShouldBindJSON(&request);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request body")); return }

	ctx := c.Request.Context()
	user, err := h.service.Create(ctx, &request)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

// Update an Existing User
func (h *UserHandler) Update(c *gin.Context) {
	var request model.UserRequest
	err := c.ShouldBindJSON(&request);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request body")); return }

	param := c.Param("id")
	userID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	if userID == sessionID { c.Error(apperror.ErrBadRequest("You cannot update your own account")); return }

	ctx := c.Request.Context()
	user, err := h.service.Update(ctx, userID, &request)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	})
}

// Delete a User
func (h *UserHandler) Delete(c *gin.Context) {
	param := c.Param("id")
	userID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	if userID == sessionID { c.Error(apperror.ErrBadRequest("You cannot delete your own account")); return }

	ctx := c.Request.Context()
	err = h.service.Delete(ctx, userID)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "User deleted successfully",
	})
}

// Get All Users
func (h *UserHandler) GetAll(c *gin.Context) {
	var filter model.UserFilterRequest
	err := c.ShouldBindQuery(&filter);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	ctx := c.Request.Context()
	users, pagination, err := h.service.GetAll(ctx, filter)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success:    true,
		Message:    "Users retrieved successfully",
		Data:       users,
		Pagination: pagination,
	})
}

// Get User by ID
func (h *UserHandler) GetByID(c *gin.Context) {
	param := c.Param("id")
	userID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	ctx := c.Request.Context()
	user, err := h.service.GetByID(ctx, userID)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// Get User by Session
func (h *UserHandler) GetBySession(c *gin.Context) {
	userID := c.GetUint("userID")
	ctx := c.Request.Context()
	user, err := h.service.GetByID(ctx, userID)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// Import Users from Excel
func (h *UserHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil { c.Error(apperror.ErrBadRequest("File is required")); return }
	if file == nil { c.Error(apperror.ErrBadRequest("File is required")); return }

	header := file.Header.Get("Content-Type")
	if header != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		c.Error(apperror.ErrBadRequest("Invalid file type, only .xlsx allowed"))
		return
	}

	excel, err := file.Open()
	if err != nil { c.Error(apperror.ErrInternal("Failed to open file")); return }
	defer excel.Close()

	ctx := c.Request.Context()
	users, err := h.service.Import(ctx, excel)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "Users imported successfully",
		Data:    users,
	})
}

// Export Users to Excel
func (h *UserHandler) Export(c *gin.Context) {
	var filter model.UserFilterRequest
	err := c.ShouldBindQuery(&filter);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid query parameters")); return }

	ctx := c.Request.Context()
	excel, err := h.service.Export(ctx, filter)
	if err != nil { c.Error(err); return }

	now := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("Export Users_%s.xlsx", now)
    c.Header("Content-Disposition", "attachment; filename=" + filename)
    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excel.Bytes())
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	param := c.Param("id")
	userID, err := utils.StringToUint(param)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid ID parameter")); return }

	sessionID := c.GetUint("userID")
	if userID == sessionID { c.Error(apperror.ErrBadRequest("You cannot reset your own password")); return }

	ctx := c.Request.Context()
	response, err := h.service.ResetPassword(ctx, userID)
	if err != nil { c.Error(err); return }

	c.JSON(200, model.Response{
		Success: true,
		Message: "User password reset successfully",
		Data:    response,
	})
}
