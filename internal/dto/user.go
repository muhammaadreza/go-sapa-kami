package dto

import (
	"time"

	"go-sapa-kami/internal/model"
)

// User Request
type UserRequest struct {
	EmployeeID uint           `json:"employee_id" binding:"required"`
	Name       string         `json:"name" binding:"required"`
	Role       model.RoleEnum `json:"role" binding:"required,oneof=user admin super_admin"`
}

// User Response
type UserResponse struct {
	ID         uint           `json:"id,omitempty"`
	EmployeeID uint           `json:"employee_id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Role       model.RoleEnum `json:"role,omitempty"`
	CreatedAt  *time.Time     `json:"created_at,omitempty"`
	UpdatedAt  *time.Time     `json:"updated_at,omitempty"`
}

// User Filter Request
type UserFilterRequest struct {
	Query     *string         `form:"query" binding:"omitempty"`
	StartDate *string         `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate   *string         `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Role      *model.RoleEnum `form:"role" binding:"omitempty,oneof=user admin super_admin"`
	Limit     int             `form:"limit" binding:"omitempty,min=1,max=100"`
	Page      int             `form:"page" binding:"omitempty,min=1"`

	SortKey   *string `form:"sort_key" binding:"omitempty,oneof=id employee_id name role created_at updated_at"`
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

