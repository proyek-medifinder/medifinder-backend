package dto

type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"error message"`
}

type PaginatedResepResponse struct {
	Data []ResepResponse `json:"data"`
	Meta MetaPagination  `json:"meta"`
}

type MetaPagination struct {
	Page      int   `json:"page" example:"1"`
	Limit     int   `json:"limit" example:"10"`
	Total     int64 `json:"total" example:"100"`
	TotalPage int64 `json:"total_page" example:"10"`
}
