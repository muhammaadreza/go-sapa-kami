package mapper

import (
	"math"

	"go-sapa-kami/internal/dto"
)

func ToPaginationResponse(count int64, page, limit int) *dto.Pagination {
	totalPages := 1
	if limit > 0 {
		if page <= 0 { page = 1 }
		totalPages = int(math.Ceil(float64(count) / float64(limit)))
		if totalPages < 1 { totalPages = 1 }
	} else {
		page = 1
		limit = int(count)
	}

	return &dto.Pagination{
		Limit:       limit,
		CurrentPage: page,
		TotalData:   int(count),
		TotalPages:  totalPages,
	}
}
