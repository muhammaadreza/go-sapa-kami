package model

// DTOs
// Auth Request
type AuthRequest struct {
	EmployeeID  uint    `json:"employee_id" binding:"required"`
	Password    string  `json:"password" binding:"required"`
	NewPassword *string `json:"new_password,omitempty" binding:"omitempty,min=8"`
}

// Token Response
type TokenResponse struct {
	Access  string `json:"access,omitempty"`
	Refresh string `json:"refresh,omitempty"`
}

// Create Token response
func CreateTokenResponse(accessToken, refreshToken string) TokenResponse {
	return TokenResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}
}
