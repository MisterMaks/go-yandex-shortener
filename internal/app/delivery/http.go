package delivery

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"encoding/json"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	ContentTypeKey     string = "Content-Type"
	TextPlainKey       string = "text/plain"
	ApplicationJSONKey string = "application/json"

	MethodKey         string = "method"
	HeaderKey         string = "header"
	RequestBodyKey    string = "request_body"
	RequestBodyStrKey string = "request_body_str"
	URLIDKey          string = "url_id"
	URLKey            string = "url"
	ShortURLKey       string = "short_url"
	RequestPathIDKey  string = "request_path_id"
	ResponseKey       string = "response"
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
	handlerLogger := logger.GetLoggerWithRequestID(r.Context())

	handlerLogger.Info("Creating or getting url")

	if r.Method != http.MethodPost {
		handlerLogger.Warn("Request method is not POST",
			zap.String(MethodKey, r.Method),
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get(ContentTypeKey)
	if !(strings.Contains(contentType, TextPlainKey) || strings.Contains(contentType, "application/x-gzip")) {
		handlerLogger.Warn("Request header \"Content-Type\" does not contain \"text/plain\" or \"application/x-gzip\"",
			zap.Any(HeaderKey, r.Header),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, TextPlainKey)))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyStr := string(body)

	url, err := ah.AppUsecase.GetOrCreateURL(bodyStr)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.String(RequestBodyStrKey, bodyStr),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := ah.AppUsecase.GenerateShortURL(url.ID)
	handlerLogger.Info("Short URL created",
		zap.String(URLIDKey, url.ID),
		zap.String(URLKey, url.URL),
		zap.String(ShortURLKey, shortURL),
	)

	w.Header().Set(ContentTypeKey, TextPlainKey)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (ah *AppHandler) APIGetOrCreateURL(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetLoggerWithRequestID(r.Context())

	handlerLogger.Info("Creating or getting url using API")

	if r.Method != http.MethodPost {
		handlerLogger.Warn("Request method is not POST",
			zap.String(MethodKey, r.Method),
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get(ContentTypeKey)
	if !(strings.Contains(contentType, ApplicationJSONKey) || strings.Contains(contentType, "application/x-gzip")) {
		handlerLogger.Warn("Request header \"Content-Type\" does not contain \"application/json\" or \"application/x-gzip\"",
			zap.Any(HeaderKey, r.Header),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, ApplicationJSONKey)))
		return
	}

	type Request struct {
		URL string `json:"url"`
	}

	var req Request
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := ah.AppUsecase.GetOrCreateURL(req.URL)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.String(URLKey, req.URL),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := ah.AppUsecase.GenerateShortURL(url.ID)
	w.Header().Set(ContentTypeKey, ApplicationJSONKey)
	w.WriteHeader(http.StatusCreated)

	type Response struct {
		Result string `json:"result"`
	}

	resp := Response{Result: shortURL}
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.Any(ResponseKey, resp),
			zap.Error(err),
		)
		return
	}
}

func (ah *AppHandler) RedirectToURL(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetLoggerWithRequestID(r.Context())

	handlerLogger.Info("Redirecting to URL")

	if r.Method != http.MethodGet {
		handlerLogger.Warn("Request method is not GET",
			zap.String(MethodKey, r.Method),
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		handlerLogger.Warn("Bad request",
			zap.String(RequestPathIDKey, id),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := ah.AppUsecase.GetURL(id)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.String(RequestPathIDKey, id),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handlerLogger.Info("Found short URL",
		zap.String(URLIDKey, url.ID),
		zap.String(URLKey, url.URL),
	)

	http.Redirect(w, r, url.URL, http.StatusTemporaryRedirect)
}
