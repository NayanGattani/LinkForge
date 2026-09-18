package server

type ShortenRequest struct {
	URL       string `json:"url"`
	Alias     string `json:"alias,omitempty"`
	ExpiresIn int64  `json:"expires_in,omitempty"`
}

type ShortenResponse struct {
	ID        string  `json:"id"`
	ShortURL  string  `json:"short_url"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type StatsResponse struct {
	ID            string  `json:"id"`
	OriginalURL   string  `json:"original_url"`
	ClickCount    int64   `json:"click_count"`
	LastClickedAt *string `json:"last_clicked_at,omitempty"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
