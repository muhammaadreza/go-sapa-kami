package handler

type Handlers struct {
	Auth      *AuthHandler
	User      *UserHandler
	Message   *MessageHandler
	Dashboard *DashboardHandler
}
