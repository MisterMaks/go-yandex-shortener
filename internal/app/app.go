package app

type URL struct {
	ID        string
	URL       string
	UserID    uint
	IsDeleted bool
}

type RequestBatchURL struct {
	CorrelationID string `json:"correlation_id"` // ID for connect OriginalURL with ShortURL in ResponseBatchURL
	OriginalURL   string `json:"original_url"`
}

type ResponseBatchURL struct {
	CorrelationID string `json:"correlation_id"` // ID for connect OriginalURL with ShortURL in RequestBatchURL
	ShortURL      string `json:"short_url"`
}

type ResponseUserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
