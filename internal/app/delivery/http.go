package delivery

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/go-chi/chi/v5"
)

const (
	ContentTypeKey string = "Content-Type"
	TextPlainKey   string = "text/plain"
)

type AppUsecaseInterface interface {
	GetOrCreateURL(rawURL string) (*app.URL, error)
	GetURL(id string) (*app.URL, error)
	GenerateShortURL(id string) string
}

type AppHandler struct {
	AppUsecase AppUsecaseInterface
}

func NewAppHandler(appUsecase AppUsecaseInterface) *AppHandler {
	return &AppHandler{AppUsecase: appUsecase}
}

func (ah *AppHandler) GetOrCreateURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	isTextPlain := false
	for _, value := range r.Header.Values(ContentTypeKey) {
		if strings.Contains(value, TextPlainKey) {
			isTextPlain = true
			break
		}
	}
	if !isTextPlain {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, TextPlainKey)))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyStr := string(body)

	url, err := ah.AppUsecase.GetOrCreateURL(bodyStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := ah.AppUsecase.GenerateShortURL(url.ID)

	w.Header().Set(ContentTypeKey, TextPlainKey)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (ah *AppHandler) RedirectToURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := ah.AppUsecase.GetURL(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url.URL, http.StatusTemporaryRedirect)
}
