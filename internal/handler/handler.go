package handler

type Handlers struct {
	Auth       *AuthHandler
	User       *UserHandler
	Aspiration *AspirationHandler
	Dashboard  *DashboardHandler
}
