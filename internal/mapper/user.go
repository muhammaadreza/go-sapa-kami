package mapper

import (
	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/model"
)

// model.User Request to model.User Model for Create
func ToUserCreateModel(request *dto.UserRequest, password string) model.User {
	return model.User{
		EmployeeID: request.EmployeeID,
		Name:       request.Name,
		Password:   password,
		Role:       request.Role,
	}
}

// model.User Request to model.User Model for Update
func ToUserUpdateModel(request *dto.UserRequest, userID uint) model.User {
	return model.User{
		ID:         userID,
		EmployeeID: request.EmployeeID,
		Name:       request.Name,
		Role:       request.Role,
	}
}

// model.User Model to model.User Response
func ToUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         user.ID,
		EmployeeID: user.EmployeeID,
		Name:       user.Name,
		Role:       user.Role,
		CreatedAt:  &user.CreatedAt,
		UpdatedAt:  &user.UpdatedAt,
	}
}

// model.User Model to model.User Nested
func ToUserNested(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		EmployeeID: user.EmployeeID,
		Role:       user.Role,
	}
}

// model.User Model to model.User Anonymous (Masked identity for discussion)
func ToUserAnonymous(user *model.User, maskID bool) dto.UserResponse {
	var id uint
	if !maskID {
		id = user.ID
	}
	return dto.UserResponse{
		ID:   id,
		Role: user.Role,
	}
}

// List Mapping Functions
// model.User Model List to model.User Response List
func ToUsersResponseList(users []model.User) []dto.UserResponse {
	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = ToUserResponse(&user)
	}

	return response
}

// model.User Model List to model.User Nested List
func ToUserNestedList(users []model.User) []dto.UserResponse {
	nested := make([]dto.UserResponse, len(users))
	for i, user := range users {
		nested[i] = ToUserNested(&user)
	}

	return nested
}

// model.User Filter Request to model.User Filter
func ToUserFilter(request *dto.UserFilterRequest) model.UserFilter {
	return model.UserFilter{
		Query:     request.Query,
		StartDate: request.StartDate,
		EndDate:   request.EndDate,
		Role:      request.Role,
		Limit:     request.Limit,
		Page:      request.Page,
		SortKey:   request.SortKey,
		SortOrder: request.SortOrder,
	}
}

// Create Token Response
func ToTokenResponse(accessToken, refreshToken string) dto.TokenResponse {
	return dto.TokenResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}
}
