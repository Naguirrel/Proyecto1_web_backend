package main

type Series struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Episodes    int    `json:"episodes"`
	Image       string `json:"image"`
}

type SeriesListResponse struct {
	Data       []Series `json:"data"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	Total      int      `json:"total"`
	TotalPages int      `json:"total_pages"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
