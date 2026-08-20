package service

import (
	"errors"
	"context"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/mapper"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/repository"

	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, request *dto.AuthRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
}

type authService struct {
	repository repository.UserRepository
	Defaults   *config.Defaults
	jwt        *config.JWT
}

func NewAuthService(repository repository.UserRepository, defaults *config.Defaults, jwt *config.JWT) AuthService {
	return &authService{repository, defaults, jwt}
}

// Login with Employee ID
func (s *authService) Login(ctx context.Context, request *dto.AuthRequest) (*dto.TokenResponse, error) {
	user, err := s.repository.GetByEmployeeID(ctx, request.EmployeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrUnauthorized("invalid credentials") }
		return nil, apperror.ErrInternal("Failed to retrieve user")
	}

	if !utils.CheckPassword(request.Password, user.Password) {
		return nil, apperror.ErrUnauthorized("invalid credentials")
	}

	if utils.CheckPassword(s.Defaults.Password, user.Password) {
		if request.NewPassword == nil { return nil, apperror.ErrConflict("default password must be changed") }

		user.Password, err = utils.HashPassword(*request.NewPassword)
		if err != nil { return nil, apperror.ErrInternal("Failed to hash password") }

		err = s.repository.Update(ctx, user)
		if err != nil { return nil, apperror.ErrInternal("Failed to update password") }
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role, s.jwt.Secret, s.jwt.AccessExpiry, s.jwt.Issuer)
	if err != nil { return nil, apperror.ErrInternal("Failed to generate access token") }

	refreshToken, err := utils.GenerateRefreshToken(user.ID, s.jwt.Secret, s.jwt.RefreshExpiry, s.jwt.Issuer)
	if err != nil { return nil, apperror.ErrInternal("Failed to generate refresh token") }

	response := mapper.ToTokenResponse(accessToken, refreshToken)
	return &response, nil
}

// Refresh Access Token using Refresh Token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	claims, err := utils.ValidateRefreshToken(refreshToken, s.jwt.Secret, s.jwt.Issuer)
	if err != nil { return nil, apperror.ErrUnauthorized("Invalid refresh token") }

	user, err := s.repository.GetByID(ctx, claims.UserID)
	if err != nil { return nil, apperror.ErrInternal("Failed to retrieve user") }

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role, s.jwt.Secret, s.jwt.AccessExpiry, s.jwt.Issuer)
	if err != nil { return nil, apperror.ErrInternal("Failed to generate access token") }

	rotatedRefreshToken, err := utils.GenerateRefreshToken(user.ID, s.jwt.Secret, s.jwt.RefreshExpiry, s.jwt.Issuer)
	if err != nil { return nil, apperror.ErrInternal("Failed to generate refresh token") }

	response := mapper.ToTokenResponse(accessToken, rotatedRefreshToken)
	return &response, nil
}
