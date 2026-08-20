package dto

import "go-sapa-kami/internal/model"

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
	CreatedAt  string         `json:"created_at,omitempty"`
	UpdatedAt  string         `json:"updated_at,omitempty"`
}

// User Filter Request
type UserFilterRequest struct {
	Query     *string         `json:"query" binding:"omitempty"`
	StartDate *string         `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate   *string         `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Role      *model.RoleEnum `json:"role" binding:"omitempty,oneof=user admin super_admin"`
	Limit     int             `json:"limit" binding:"omitempty,min=1,max=100"`
	Page      int             `json:"page" binding:"omitempty,min=1"`

	SortKey   *string `json:"sort_key" binding:"omitempty,oneof=id employee_id name role created_at updated_at"`
	SortOrder *string `json:"sort_order" binding:"omitempty,oneof=asc desc"`
}

