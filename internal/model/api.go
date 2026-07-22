package model

import "math"

// Response structure
type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Detail	   any         `json:"detail,omitempty"`
	Data       any 	   	   `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination response
type Pagination struct {
	Limit   	int `json:"limit"`
	TotalData   int `json:"total_data"`
    CurrentPage int `json:"current_page"`
    TotalPages  int `json:"total_pages"`
}

// Mapping Functions
// Create Pagination response
func CreatePagination(count int64, page, limit int) *Pagination {
	totalPages := 1
	if limit > 0 {
        if page <= 0 { page = 1 }
        totalPages = int(math.Ceil(float64(count) / float64(limit)))
    } else {
		page = 1
        limit = int(count)
    }

	return &Pagination{
		Limit:       limit,
		CurrentPage: page,
		TotalData:   int(count),
		TotalPages:  totalPages,
	}
}
