package main

import (
	"log"
	"net/http"

	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	appRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
	appUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
)

const (
	Addr                          = "127.0.0.1:8080"
	CountRegenerationsForLengthID = 5
	LengthID                      = 5
	MaxLengthID                   = 20
)

func main() {
	appRepo := appRepoInternal.NewAppRepo()
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
	mux.HandleFunc(`/`, appHandler.CreateOrGet)

	log.Printf("INFO\tServer running on %s ...\n", Addr)
	err = http.ListenAndServe(Addr, mux)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to start server. Error:", err)
	}
}
