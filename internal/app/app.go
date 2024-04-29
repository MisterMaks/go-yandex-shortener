package app

type URL struct {
	ID  string
	URL string
}

func NewURL(id, url string) *URL {
	return &URL{ID: id, URL: url}
}
