package service

import (
	"io"
	"bytes"
	"errors"
	"context"

	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/repository"

	"gorm.io/gorm"
)

type UserService interface {
	Create(ctx context.Context, request *model.UserRequest) (*model.UserResponse, error)
	Update(ctx context.Context, userID uint, request *model.UserRequest) (*model.UserResponse, error)
	Delete(ctx context.Context, userID uint) error
	GetAll(ctx context.Context, filter model.UserFilterRequest) ([]model.UserResponse, *model.Pagination, error)
	GetByID(ctx context.Context, userID uint) (*model.UserResponse, error)

	Count(ctx context.Context, filter model.UserFilterRequest) (int64, error)
	CountDaily(ctx context.Context, filter model.UserFilterRequest) (map[string]int64, error)
	ResetPassword(ctx context.Context, userID uint) (*model.UserResponse, error)
	Import(ctx context.Context, excel io.Reader) ([]model.UserResponse, error)
	Export(ctx context.Context, filter model.UserFilterRequest) (*bytes.Buffer, error)
}

type userService struct {
	repository  repository.UserRepository
	defaults	*config.Defaults
}

func NewUserService(repository repository.UserRepository, defaults *config.Defaults) UserService {
	return &userService{ repository, defaults }
}

// Create a New User
func (s *userService) Create(ctx context.Context, request *model.UserRequest) (*model.UserResponse, error) {
	password, err := utils.HashPassword(s.defaults.Password)
	if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

	user := request.ToCreateModel(password)
	err = s.repository.Create(ctx, &user)
	if err != nil { return nil, apperror.ErrInternal("Failed to create user") }

	response := user.ToResponse()
	return &response, nil
}

// Update an Existing User
func (s *userService) Update(ctx context.Context, userID uint, request *model.UserRequest) (*model.UserResponse, error) {
	user := request.ToUpdateModel(&userID)
	err := s.repository.Update(ctx, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to update user")
	}

	response := user.ToResponse()
	return &response, nil
}

// Delete a User
func (s *userService) Delete(ctx context.Context, userID uint) error {
	err := s.repository.Delete(ctx, userID)
	if err != nil { return apperror.ErrInternal("Failed to delete user") }

	return nil
}

// Get All Users
func (s *userService) GetAll(ctx context.Context, filter model.UserFilterRequest) ([]model.UserResponse, *model.Pagination, error) {
	users, count, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, nil, apperror.ErrInternal("Failed to retrieve users") }

	response := model.ToUsersResponseList(users)
	pagination := model.CreatePagination(count, filter.Page, filter.Limit)
	return response, pagination, nil
}

// Get User by ID
func (s *userService) GetByID(ctx context.Context, userID uint) (*model.UserResponse, error) {
	user, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to retrieve user")
	}

	response := user.ToResponse()
	return &response, nil
}

// Count Users
func (s *userService) Count(ctx context.Context, filter model.UserFilterRequest) (int64, error) {
	count, err := s.repository.Count(ctx, filter)
	if err != nil { return 0, apperror.ErrInternal("Failed to count users") }

	return count, nil
}

// Count Daily Users
func (s *userService) CountDaily(ctx context.Context, filter model.UserFilterRequest) (map[string]int64, error) {
	counts, err := s.repository.CountDaily(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to count daily users") }

	return counts, nil
}

// Reset User Password to Default
func (s *userService) ResetPassword(ctx context.Context, userID uint) (*model.UserResponse, error) {
	password, err := utils.HashPassword(s.defaults.Password)
	if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

	user := model.User{ ID: userID, Password: password }
	err = s.repository.Update(ctx, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to reset password")
	}

	response := user.ToResponse()
	return &response, nil
}

// Create Many Users from Excel File
func (s *userService) Import(ctx context.Context, excel io.Reader) ([]model.UserResponse, error) {
	requests, err := utils.ParseUsersExcel(excel)
	if err != nil { return nil, apperror.ErrInternal("Failed to parse users from Excel file") }
	if len(requests) == 0 { return nil, apperror.ErrBadRequest("No users found in Excel file") }

	password, err := utils.HashPassword(s.defaults.Password)
	if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

	var users []model.User
	for _, request := range requests {
		user := request.ToCreateModel(password)
		users = append(users, user)
	}

	err = s.repository.CreateMany(ctx, &users)
	if err != nil { return nil, apperror.ErrInternal("Failed to create users") }

	response := model.ToUsersResponseList(users)
	return response, nil
}

// Generate Users Excel
func (s *userService) Export(ctx context.Context, filter model.UserFilterRequest) (*bytes.Buffer, error) {
	users, _, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to retrieve users") }

	excel, err := utils.GenerateUsersExcel(users)
	if err != nil { return nil, apperror.ErrInternal("Failed to generate Excel file") }

	return excel, nil
}
