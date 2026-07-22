package model

import "time"

// Role Enums
type RoleEnum string

// Constants for Role enums
const (
	RoleUser       RoleEnum = "user"
	RoleAdmin 	   RoleEnum = "admin"
	RoleSuperAdmin RoleEnum = "super_admin"
)

// Users Model
type User struct {
	ID		   uint	     `gorm:"primaryKey"`
	EmployeeID uint	     `gorm:"unique;not null"`
	Name	   string    `gorm:"not null"`
	Password   string    `gorm:"not null"`

	Role	   RoleEnum  `gorm:"type:role_enum;default:'user';not null"`
	CreatedAt  time.Time `gorm:"<-:create"`
    UpdatedAt  time.Time `gorm:"autoUpdateTime"`

	Messages   []Message `gorm:"foreignKey:UserID"`
}

// DTOs
// User Request
type UserRequest struct {
	Name 		string   `json:"name" binding:"required"`
	EmployeeID  uint     `json:"employee_id" binding:"required"`
	Role 		RoleEnum `json:"role" binding:"required,oneof=user admin super_admin"`
}

// User Response
type UserResponse struct {
	ID    		uint     `json:"id"`
	EmployeeID  uint     `json:"employee_id"`
	Name  		string   `json:"name"`
	Role        RoleEnum `json:"role,omitempty"`

	CreatedAt 	string  `json:"created_at,omitempty"`
	UpdatedAt 	string  `json:"updated_at,omitempty"`
}

// User Filter Request
type UserFilterRequest struct {
	Query     *string    `form:"query" binding:"omitempty"`
	StartDate *string   `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
    EndDate   *string   `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Role      *RoleEnum `form:"role" binding:"omitempty,oneof=user admin super_admin"`
	Limit     int       `form:"limit" binding:"omitempty,min=1,max=100"`
	Page      int       `form:"page" binding:"omitempty,min=1"`

	SortKey   *string    `form:"sort_key" binding:"omitempty,oneof=id employee_id name role created_at updated_at"`
	SortOrder *string    `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// Mapping Functions
// User Request to User Model for Create
func (request *UserRequest) ToCreateModel(password string) User {
	return User{
		EmployeeID: request.EmployeeID,
		Name: 		request.Name,
		Password: 	password,
		Role: 		request.Role,
	}
}

// User Request to User Model for Update
func (request *UserRequest) ToUpdateModel(userID *uint) User {
	return User{
		ID:   		*userID,
		EmployeeID: request.EmployeeID,
		Name: 		request.Name,
		Role: 		request.Role,
	}
}

// User Model to User Response
func (user *User) ToResponse() UserResponse {
	return UserResponse{
		ID: 		 user.ID,
		EmployeeID:  user.EmployeeID,
		Name: 		 user.Name,
		Role: 		 user.Role,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// User Model to User Nested
func (user *User) ToNested() UserResponse {
	return UserResponse{
		ID:         user.ID,
		Name: 		user.Name,
		EmployeeID: user.EmployeeID,
	}
}

// List Mapping Functions
// User Model List to User Response List
func ToUsersResponseList(users []User) []UserResponse {
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = user.ToResponse()
	}

	return response
}

// User Model List to User Nested List
func ToUserNestedList(users []User) []UserResponse {
	nested := make([]UserResponse, len(users))
	for i, user := range users {
		nested[i] = user.ToNested()
	}

	return nested
}
