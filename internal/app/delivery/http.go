package delivery

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
)

const (
	ContentTypeKey string = "Content-Type"
	TextPlainKey   string = "text/plain"
)

type AppUsecaseInterface interface {
	Create(s string) (*app.URL, error)
	Get(id string) (*app.URL, error)
	GenerateShortURL(addr, id string) string
}

type AppHandler struct {
	AppUsecase AppUsecaseInterface
}

func NewAppHandler(appUsecase AppUsecaseInterface) *AppHandler {
	return &AppHandler{AppUsecase: appUsecase}
}

func (ah *AppHandler) CreateOrGet(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO\tAppHandler.CreateOrGet()")

	uri := r.RequestURI
	uriS := strings.Split(uri, "/")
	uriSLength := len(uriS)
	switch {
	case uriSLength == 2 && uriS[1] == "":
		ah.Create(w, r)
	case uriSLength == 2 && uriS[1] != "":
		ah.Get(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (ah *AppHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO\tAppHandler.Create()")

	if r.Method != http.MethodPost {
		log.Printf("WARNING\tBad request method. Method: %s\n", r.Method)
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
		log.Printf("WARNING\tBad request header %s. %s: %s\n", ContentTypeKey, ContentTypeKey, r.Header.Get(ContentTypeKey))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header %s is not %s", ContentTypeKey, TextPlainKey)))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("WARNING\tBad request. Request body: %s\n", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyStr := string(body)

	url, err := ah.AppUsecase.Create(bodyStr)
	if err != nil {
		log.Printf("WARNING\tBad request. Request body string: %s. Error: %v\n", bodyStr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := ah.AppUsecase.GenerateShortURL(r.Host, url.ID)
	log.Printf("INFO\tURL ID: %s, URL: %s, short URL: %s\n", url.ID, url.URL, shortURL)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (ah *AppHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO\tAppHandler.Get()")

	if r.Method != http.MethodGet {
		log.Printf("WARNING\tBad request method. Method: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	uri := r.RequestURI
	uriS := strings.Split(uri, "/")
	if len(uriS) != 2 {
		log.Printf("WARNING\tBad request URI. Request URI: %s\n", uri)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := uriS[1]
	if id == "" {
		log.Printf("WARNING\tBad request. Request path id: %s\n", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := ah.AppUsecase.Get(id)
	if err != nil {
		log.Printf("WARNING\tBad request. Request path id: %s. Error: %v\n", id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("INFO\tURL ID: %s, URL: %s\n", url.ID, url.URL)

	http.Redirect(w, r, url.URL, http.StatusTemporaryRedirect)
}
