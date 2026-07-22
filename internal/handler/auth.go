package handler

import (
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/service"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
	app     *config.App
	jwt	    *config.JWT
}

func NewAuthHandler(service service.AuthService, config *config.Config) *AuthHandler {
	return &AuthHandler{ service: service, app: config.App, jwt: config.JWT }
}

// Login
func (h *AuthHandler) Login(c *gin.Context) {
	var request model.AuthRequest
	err := c.ShouldBindJSON(&request);
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request")); return }

	ctx := c.Request.Context()
	token, err := h.service.Login(ctx, &request)
	if err != nil { c.Error(err); return }

	c.SetCookie("access_token", token.Access, h.jwt.AccessMaxAge, "/", h.app.Domain, false, true)
	c.SetCookie("refresh_token", token.Refresh, h.jwt.RefreshMaxAge, "/", h.app.Domain, false, true)
	c.JSON(200, model.Response{
		Success: true,
		Message: "Login successfully",
		Data:    token,
	})
}

// Refresh Access Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil { c.Error(apperror.ErrUnauthorized("Refresh token not found")); return }

	ctx := c.Request.Context()
	token, err := h.service.RefreshToken(ctx, refreshToken)
	if err != nil { c.Error(err); return }

	c.SetCookie("access_token", token.Access, h.jwt.AccessMaxAge, "/", h.app.Domain, false, true)
	c.JSON(200, model.Response{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    token,
	})
}

// Logout user by clearing cookies
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", h.app.Domain, false, true)
	c.SetCookie("refresh_token", "", -1, "/", h.app.Domain, false, true)
	c.JSON(200, model.Response{
		Success: true,
		Message: "Logged out successfully",
		Data:    nil,
	})
}

// Microsoft OAuth2 Handlers (draft)
// Generate Login URL
// func (h *AuthHandler) GenerateLoginURL(c *gin.Context) {
// 	auth, err := h.service.GenerateLoginURL()
// 	if err != nil { c.Error(err); return }

// 	c.SetCookie("oauth_state", auth.State, 3600, "/", h.app.Domain, false, true)
// 	c.JSON(200, model.Response{
// 		Success: true,
// 		Message: "Login URL generated successfully",
// 		Data:    auth,
// 	})
// }

// Exchange Code to Access and Refresh Tokens
// func (h *AuthHandler) Callback(c *gin.Context) {
// 	var callback model.AzureCallback
// 	err := c.ShouldBindQuery(&callback);
// 	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request")); return }

// 	state, err := c.Cookie("oauth_state")
// 	c.SetCookie("oauth_state", "", -1, "/", h.app.Domain, false, true)
// 	if err != nil || state != callback.State { c.Error(apperror.ErrUnauthorized("Invalid state parameter")); return }

// 	ctx := c.Request.Context()
// 	token, err := h.service.ExchangeToken(ctx, callback.Code)
// 	if err != nil { c.Error(err); return }

// 	c.SetCookie("access_token", token.Access, h.jwt.AccessMaxAge, "/", h.app.Domain, false, true)
// 	c.SetCookie("refresh_token", token.Refresh, h.jwt.RefreshMaxAge, "/", h.app.Domain, false, true)
// 	c.JSON(200, model.Response{
// 		Success: true,
// 		Message: "Login successfully",
// 		Data:    token,
// 	})
// }
