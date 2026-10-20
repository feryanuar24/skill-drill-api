package models

import "math"

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func CalculatePagination(totalItems int64, page, limit int) Pagination {
	if page < 1 {
		page = 1
	}
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(limit)))
	} else if totalItems > 0 {
		totalPages = 1
		limit = int(totalItems)
	}

	return Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

