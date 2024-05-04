package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	appRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
	appUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
)

const (
	Addr                          string = ":8080"
	CountRegenerationsForLengthID uint   = 5
	LengthID                      uint   = 5
	MaxLengthID                   uint   = 20
)

type AppHandlerInterface interface {
	GetOrCreateURL(w http.ResponseWriter, r *http.Request)
	RedirectToURL(w http.ResponseWriter, r *http.Request)
}

func ShortenerRouter(appHandler AppHandlerInterface) chi.Router {
	r := chi.NewRouter()
	r.Route(`/`, func(r chi.Router) {
		r.Post(`/`, appHandler.GetOrCreateURL)
		r.Get(`/{id}`, appHandler.RedirectToURL)
	})
	return r
}

func main() {
	appRepo := appRepoInternal.NewAppRepoInmem()
	appUsecase, err := appUsecaseInternal.NewAppUsecase(
		appRepo,
		CountRegenerationsForLengthID,
		LengthID,
		MaxLengthID,
	)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to create appUsecase. Error:", err)
	}

	appHandler := appDeliveryInternal.NewAppHandler(appUsecase)

	r := ShortenerRouter(appHandler)

	log.Printf("INFO\tServer running on %s ...\n", Addr)
	err = http.ListenAndServe(Addr, r)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to start server. Error:", err)
	}
}
