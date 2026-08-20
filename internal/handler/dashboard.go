package handler

import (
	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service}
}

// Get dashboard data
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	response, err := h.service.GetDashboard(ctx)
	if err != nil { c.Error(err); return }

	c.JSON(200, dto.Response{
		Success: true,
		Message: "Dashboard data retrieved successfully",
		Data:    response,
	})
}
