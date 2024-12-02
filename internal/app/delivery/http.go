package delivery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Constants for http handlers.
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
	URLsKey           string = "urls"
	ShortURLKey       string = "short_url"
	RequestPathIDKey  string = "request_path_id"
	ResponseKey       string = "response"
)

// AppUsecaseInterface contains the necessary functions for the business logic of app.
type AppUsecaseInterface interface {
	GetOrCreateURL(rawURL string, userID uint) (*app.URL, bool, error)                                   // get created or create short URL for request URL
	GetURL(id string) (*app.URL, error)                                                                  // get original URL for short URL
	GenerateShortURL(id string) string                                                                   // generate short URL
	Ping() error                                                                                         // ping database
	GetOrCreateURLs(requestBatchURLs []app.RequestBatchURL, userID uint) ([]app.ResponseBatchURL, error) // get created or create short URLs for request batch URLs
	GetUserURLs(userID uint) ([]app.ResponseUserURL, error)                                              // get short and original URLs for user
	SendDeleteUserURLsInChan(userID uint, urlIDs []string)                                               // send urls in delete chan
}

// AppHandler handlers struct.
type AppHandler struct {
	AppUsecase AppUsecaseInterface
}

// NewAppHandler creates *AppHandler
func NewAppHandler(appUsecase AppUsecaseInterface) *AppHandler {
	return &AppHandler{AppUsecase: appUsecase}
}

// GetOrCreateURL Get (if URL existed) or create URL.
//
//	@Summary	Get (if URL existed) or create URL
//	@Accept		plain
//	@Produce	plain
//	@Param		url	body		string	true	"URL"	example(https://test.org)
//	@Success	201	{string}	string	"URL created"
//	@Success	409	{string}	string	"URL exists"
//	@Failure	405	{string}	string	"Method not allowed"
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Router		/ [post]
func (ah *AppHandler) GetOrCreateURL(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

	handlerLogger.Info("Creating or getting URL")

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

	userID, err := usecase.GetContextUserID(r.Context())
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bodyStr := string(body)

	url, exists, err := ah.AppUsecase.GetOrCreateURL(bodyStr, userID)
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
	if exists {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write([]byte(shortURL))
}

// APIGetOrCreateURL Get (if URL existed) or create URL in JSON format.
//
//	@Summary	Get (if URL existed) or create URL in JSON format
//	@Accept		json
//	@Produce	json
//	@Param		url	body		delivery.APIGetOrCreateURL.Request	true	"URL"
//	@Success	201	{object}	delivery.APIGetOrCreateURL.Response	"URL created"
//	@Success	409	{object}	delivery.APIGetOrCreateURL.Response	"URL exists"
//	@Failure	405	{string}	string								"Method not allowed"
//	@Failure	400	{string}	string								"Bad request"
//	@Failure	401	{string}	string								"Unauthorized"
//	@Router		/api/shorten [post]
func (ah *AppHandler) APIGetOrCreateURL(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

	handlerLogger.Info("Creating or getting URL using API")

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

	userID, err := usecase.GetContextUserID(r.Context())
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	url, exists, err := ah.AppUsecase.GetOrCreateURL(req.URL, userID)
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
	if exists {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

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

// RedirectToURL Redirect to original URL.
//
//	@Summary	Redirect to original URL
//	@Produce	html
//	@Param		url_id	path		string	true	"URL ID"	example(qwerty)
//	@Success	307		{body}		string	"Redirect to original URL"
//	@Failure	405		{string}	string	"Method not allowed"
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	410		{string}	string	"Gone"
//	@Router		/{url_id} [get]
func (ah *AppHandler) RedirectToURL(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

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

	handlerLogger.Info("Found URL",
		zap.Any(URLKey, url),
	)

	if url.IsDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	http.Redirect(w, r, url.URL, http.StatusTemporaryRedirect)
}

// Ping Ping database.
//
//	@Summary	Ping database
//	@Produce	plain
//	@Success	200	{string}	string	"OK"
//	@Failure	405	{string}	string	"Method not allowed"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/ping [get]
func (ah *AppHandler) Ping(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

	handlerLogger.Info("Ping DB")

	if r.Method != http.MethodGet {
		handlerLogger.Warn("Request method is not GET",
			zap.String(MethodKey, r.Method),
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := ah.AppUsecase.Ping()
	if err != nil {
		handlerLogger.Error("Failed to ping DB",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

// APIGetOrCreateURLs Get (if URLs existed) or create URLs in JSON format.
//
//	@Summary	Get (if URLs existed) or create URLs in JSON format
//	@Accept		json
//	@Produce	json
//	@Param		url	body		[]app.RequestBatchURL	true	"URL"
//	@Success	201	{object}	[]app.ResponseBatchURL	"URLs created"
//	@Failure	405	{string}	string					"Method not allowed"
//	@Failure	400	{string}	string					"Bad request"
//	@Failure	401	{string}	string					"Unauthorized"
//	@Router		/api/shorten/batch [post]
func (ah *AppHandler) APIGetOrCreateURLs(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

	handlerLogger.Info("Creating or getting URLs batch using API")

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

	var req []app.RequestBatchURL
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

	userID, err := usecase.GetContextUserID(r.Context())
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := ah.AppUsecase.GetOrCreateURLs(req, userID)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.Any(URLsKey, req),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set(ContentTypeKey, ApplicationJSONKey)
	w.WriteHeader(http.StatusCreated)

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

// APIGetUserURLs Get user URLs in JSON format.
//
//	@Summary	Get user URLs in JSON format
//	@Produce	json
//	@Success	200	{object}	[]app.ResponseUserURL	"URLs created"
//	@Failure	405	{string}	string					"Method not allowed"
//	@Failure	400	{string}	string					"Bad request"
//	@Failure	401	{string}	string					"Unauthorized"
//	@Failure	204	{string}	string					"No content"
//	@Router		/api/user/urls [get]
func (ah *AppHandler) APIGetUserURLs(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

	handlerLogger.Info("Getting user URLs using API")

	if r.Method != http.MethodGet {
		handlerLogger.Warn("Request method is not GET", zap.String(MethodKey, r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := usecase.GetContextUserID(r.Context())
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := ah.AppUsecase.GetUserURLs(userID)
	if err != nil {
		handlerLogger.Warn("Bad request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(resp) == 0 {
		handlerLogger.Warn("No content")
		w.WriteHeader(http.StatusNoContent)
	}

	w.Header().Set(ContentTypeKey, ApplicationJSONKey)

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

// APIDeleteUserURLs Delete user URLs in JSON format.
//
//	@Summary	Delete user URLs in JSON format
//	@Accept		json
//	@Produce	plain
//	@Param		url_ids	body		[]string	true	"URL IDs"
//	@Success	202		{string}	string		"Accepted"
//	@Failure	405		{string}	string		"Method not allowed"
//	@Failure	400		{string}	string		"Bad request"
//	@Failure	401		{string}	string		"Unauthorized"
//	@Security	ApiKeyAuth
//	@Router		/api/user/urls [delete]
func (ah *AppHandler) APIDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	handlerLogger := logger.GetContextLogger(r.Context())

	handlerLogger.Info("Deleting user URLs using API")

	if r.Method != http.MethodDelete {
		handlerLogger.Warn("Request method is not DELETE", zap.String(MethodKey, r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req []string
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

	handlerLogger.Debug("Request data", zap.Any("url_ids", req))

	userID, err := usecase.GetContextUserID(r.Context())
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ah.AppUsecase.SendDeleteUserURLsInChan(userID, req)

	w.WriteHeader(http.StatusAccepted)
}
