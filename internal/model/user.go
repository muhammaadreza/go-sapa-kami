package model

import "time"

// Role Enums
type RoleEnum string

// Constants for Role enums
const (
	RoleUser       RoleEnum = "user"
	RoleAdmin      RoleEnum = "admin"
	RoleSuperAdmin RoleEnum = "super_admin"
)

// Users Model
type User struct {
	ID   	   uint          `gorm:"primaryKey"`
	Name 	   string        `gorm:"not null"`
	Role 	   RoleEnum      `gorm:"type:role_enum;default:'user';not null"`

	EmployeeID uint          `gorm:"unique;not null"`
	Password   string        `gorm:"not null"`

	CreatedAt  time.Time     `gorm:"<-:create"`
	UpdatedAt  time.Time     `gorm:"autoUpdateTime"`

	Aspirations []Aspiration `gorm:"foreignKey:SenderID"`
}

// User Filter
type UserFilter struct {
	Query     *string
	StartDate *string
	EndDate   *string
	Role      *RoleEnum
	Limit     int
	Page      int

	SortKey   *string
	SortOrder *string
}
