package app

// URL struct for URL.
type URL struct {
	ID        string
	URL       string
	UserID    uint
	IsDeleted bool
}

// RequestBatchURL struct for APIGetOrCreateURLs handler.
type RequestBatchURL struct {
	CorrelationID string `json:"correlation_id"` // ID for connect OriginalURL with ShortURL in ResponseBatchURL
	OriginalURL   string `json:"original_url"`
}

// ResponseBatchURL struct for APIGetOrCreateURLs handler.
type ResponseBatchURL struct {
	CorrelationID string `json:"correlation_id"` // ID for connect OriginalURL with ShortURL in RequestBatchURL
	ShortURL      string `json:"short_url"`
}

// ResponseUserURL struct for APIGetUserURLs handler.
type ResponseUserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
