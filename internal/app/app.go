package app

type URL struct {
	ID  string
	URL string
}

type RequestBatchURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatchURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
