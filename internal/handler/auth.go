package handler

import (
	"net/http"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/service"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
	app     *config.App
	jwt     *config.JWT
}

func NewAuthHandler(service service.AuthService, config *config.Config) *AuthHandler {
	return &AuthHandler{service: service, app: config.App, jwt: config.JWT}
}

// Login
func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.AuthRequest
	err := c.ShouldBindJSON(&request)
	if err != nil { c.Error(apperror.ErrBadRequest("Invalid request")); return }

	ctx := c.Request.Context()
	token, err := h.service.Login(ctx, &request)
	if err != nil { c.Error(err); return }

	h.setAuthCookie(c, "access_token", token.Access, h.jwt.AccessMaxAge)
	h.setAuthCookie(c, "refresh_token", token.Refresh, h.jwt.RefreshMaxAge)
	c.JSON(200, dto.Response{
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

	h.setAuthCookie(c, "access_token", token.Access, h.jwt.AccessMaxAge)
	h.setAuthCookie(c, "refresh_token", token.Refresh, h.jwt.RefreshMaxAge)
	c.JSON(200, dto.Response{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    token,
	})
}

// Logout user by clearing cookies
func (h *AuthHandler) Logout(c *gin.Context) {
	h.setAuthCookie(c, "access_token", "", -1)
	h.setAuthCookie(c, "refresh_token", "", -1)
	c.JSON(200, dto.Response{
		Success: true,
		Message: "Logged out successfully",
		Data:    nil,
	})
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, name, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   h.app.CookieDomain,
		MaxAge:   maxAge,
		Secure:   h.app.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
