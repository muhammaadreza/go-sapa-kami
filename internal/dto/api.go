package dto

// API Response
type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Detail     any         `json:"detail,omitempty"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination response
type Pagination struct {
	Limit       int `json:"limit"`
	TotalData   int `json:"total_data"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}
