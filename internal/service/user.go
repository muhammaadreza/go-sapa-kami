package service

import (
	"io"
	"log"
	"bytes"
	"errors"
	"context"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/mapper"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/repository"

	"gorm.io/gorm"
)

type UserService interface {
	// CRUD Operations
	Create(ctx context.Context, request *dto.UserRequest) (*dto.UserResponse, error)
	Update(ctx context.Context, userID uint, request *dto.UserRequest) (*dto.UserResponse, error)
	Delete(ctx context.Context, userID uint) error
	GetAll(ctx context.Context, request *dto.UserFilterRequest) ([]dto.UserResponse, *dto.Pagination, error)
	GetByID(ctx context.Context, userID uint) (*dto.UserResponse, error)
	ResetPassword(ctx context.Context, userID uint) (*dto.UserResponse, error)

	// Count Operations
	Count(ctx context.Context, request *dto.UserFilterRequest) (int64, error)
	CountActive(ctx context.Context, request *dto.UserFilterRequest) (int64, error)
	CountDaily(ctx context.Context, request *dto.UserFilterRequest) (map[string]int64, error)

	// Import and Export Operations
	Import(ctx context.Context, excel io.Reader) ([]dto.UserResponse, error)
	Export(ctx context.Context, request *dto.UserFilterRequest) (*bytes.Buffer, error)
}

type userService struct {
	repository repository.UserRepository
	defaults   *config.Defaults
}

func NewUserService(repository repository.UserRepository, defaults *config.Defaults) UserService {
	return &userService{repository, defaults}
}

// Create a New User
func (s *userService) Create(ctx context.Context, request *dto.UserRequest) (*dto.UserResponse, error) {
	password, err := utils.HashPassword(s.defaults.Password)
	if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

	user := mapper.ToUserCreateModel(request, password)
	err = s.repository.Create(ctx, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) { return nil, apperror.ErrConflict("Employee ID already exists") }
		return nil, apperror.ErrInternal("Failed to create user")
	}

	response := mapper.ToUserResponse(&user)
	return &response, nil
}

// Update an Existing User
func (s *userService) Update(ctx context.Context, userID uint, request *dto.UserRequest) (*dto.UserResponse, error) {
	user := mapper.ToUserUpdateModel(request, userID)
	err := s.repository.Update(ctx, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to update user")
	}

	response := mapper.ToUserResponse(&user)
	return &response, nil
}

// Delete a User
func (s *userService) Delete(ctx context.Context, userID uint) error {
	attachmentPaths, err := s.repository.GetAttachmentPaths(ctx, userID)
	if err != nil { return apperror.ErrInternal("Failed to retrieve user attachments") }

	err = s.repository.Delete(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return apperror.ErrNotFound("User not found") }
		return apperror.ErrInternal("Failed to delete user")
	}

	err = utils.DeleteFiles(attachmentPaths);
	if err != nil { log.Printf("failed to delete attachments for user %d: %v", userID, err) }

	return nil
}

// Get All Users
func (s *userService) GetAll(ctx context.Context, request *dto.UserFilterRequest) ([]dto.UserResponse, *dto.Pagination, error) {
	filter := mapper.ToUserFilter(request)
	users, count, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, nil, apperror.ErrInternal("Failed to retrieve users") }

	response := mapper.ToUsersResponseList(users)
	pagination := mapper.ToPaginationResponse(count, filter.Page, filter.Limit)
	return response, pagination, nil
}

// Get User by ID
func (s *userService) GetByID(ctx context.Context, userID uint) (*dto.UserResponse, error) {
	user, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to retrieve user")
	}

	response := mapper.ToUserResponse(user)
	return &response, nil
}

// Reset User Password to Default
func (s *userService) ResetPassword(ctx context.Context, userID uint) (*dto.UserResponse, error) {
	password, err := utils.HashPassword(s.defaults.Password)
	if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

	user := model.User{ ID: userID, Password: password }
	err = s.repository.Update(ctx, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to reset password")
	}

	response := mapper.ToUserResponse(&user)
	return &response, nil
}

// Count Users
func (s *userService) Count(ctx context.Context, request *dto.UserFilterRequest) (int64, error) {
	filter := mapper.ToUserFilter(request)
	count, err := s.repository.Count(ctx, filter)
	if err != nil { return 0, apperror.ErrInternal("Failed to count users") }

	return count, nil
}

// Count Active Users
func (s *userService) CountActive(ctx context.Context, request *dto.UserFilterRequest) (int64, error) {
	filter := mapper.ToUserFilter(request)
	count, err := s.repository.CountActive(ctx, filter)
	if err != nil { return 0, apperror.ErrInternal("Failed to count active users") }

	return count, nil
}

// Count Daily Users
func (s *userService) CountDaily(ctx context.Context, request *dto.UserFilterRequest) (map[string]int64, error) {
	filter := mapper.ToUserFilter(request)
	counts, err := s.repository.CountDaily(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to count daily users") }

	return counts, nil
}

// Create Many Users from Excel File
func (s *userService) Import(ctx context.Context, excel io.Reader) ([]dto.UserResponse, error) {
	requests, err := utils.ParseUsersExcel(excel)
	if err != nil { return nil, apperror.ErrBadRequest(err.Error()) }
	if len(requests) == 0 { return nil, apperror.ErrBadRequest("No users found in Excel file") }

	password, err := utils.HashPassword(s.defaults.Password)
	if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

	var users []model.User
	for _, request := range requests {
		user := mapper.ToUserCreateModel(&request, password)
		users = append(users, user)
	}

	err = s.repository.CreateMany(ctx, &users)
	if err != nil { return nil, apperror.ErrInternal("Failed to create users") }

	response := mapper.ToUsersResponseList(users)
	return response, nil
}

// Generate Users Excel
func (s *userService) Export(ctx context.Context, request *dto.UserFilterRequest) (*bytes.Buffer, error) {
	filter := mapper.ToUserFilter(request)
	filter.Limit = 0
	filter.Page = 0

	users, _, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to retrieve users") }

	excel, err := utils.GenerateUsersExcel(users)
	if err != nil { return nil, apperror.ErrInternal("Failed to generate Excel file") }

	return excel, nil
}
