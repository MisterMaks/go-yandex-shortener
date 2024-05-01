package main

import (
	"log"
	"net/http"

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

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, appHandler.Create)
	mux.HandleFunc(`/{id}`, appHandler.Get)

	log.Printf("INFO\tServer running on %s ...\n", Addr)
	err = http.ListenAndServe(Addr, mux)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to start server. Error:", err)
	}
}
