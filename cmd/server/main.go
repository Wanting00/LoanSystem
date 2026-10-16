package main

import (
	"errors"
	"log"
	"net/http"

	"loan-service/config"
	"loan-service/internal/handler"
	"loan-service/internal/service"
)

func main() {
	cfg := config.Load()
	router := handler.NewRouter(service.NewLoanService())

	log.Printf("loan service listening on %s", cfg.HTTPAddress)
	if err := http.ListenAndServe(cfg.HTTPAddress, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
