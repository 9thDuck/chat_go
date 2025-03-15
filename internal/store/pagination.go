package store

import "math"

type Pagination struct {
	Limit         int    `json:"limit"`
	Page          int    `json:"page"`
	Sort          string `json:"sort"`
	SortDirection string `json:"sort_direction"`
}

func (p *Pagination) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

func (p *Pagination) CalculateTotalPages(total int) int {
	return int(math.Ceil(float64(total) / float64(p.Limit)))
}

type PaginatedResult struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
}
