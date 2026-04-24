package main

type Series struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Episodes      int     `json:"episodes"`
	Image         string  `json:"image"`
	RatingAverage float64 `json:"rating_average"`
	RatingCount   int     `json:"rating_count"`
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

type Rating struct {
	ID        int    `json:"id"`
	SeriesID  int    `json:"series_id"`
	Reviewer  string `json:"reviewer"`
	Score     int    `json:"score"`
	CreatedAt string `json:"created_at"`
}

type RatingInput struct {
	Reviewer string `json:"reviewer"`
	Score    int    `json:"score"`
}

type RatingResponse struct {
	SeriesID      int      `json:"series_id"`
	Average       float64  `json:"average"`
	Count         int      `json:"count"`
	LatestRatings []Rating `json:"latest_ratings"`
}
