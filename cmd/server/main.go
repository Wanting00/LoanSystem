package main

import (
	"errors"
	"log"
	"net/http"

	"loan-service/config"
	"loan-service/internal/handler"
	"loan-service/internal/repository"
	"loan-service/internal/service"
)

func main() {
	cfg := config.Load()
	db, err := config.OpenMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}

	loanRepository := repository.NewLoanRepository(db)
	borrowerService := service.NewBorrowerService(repository.NewBorrowerRepository(db))
	loanService := service.NewLoanService(db, loanRepository, borrowerService)
	investService := service.NewInvestService(
		db,
		loanRepository,
		repository.NewInvestmentRepository(),
		repository.NewInvestorRepository(),
		service.NewNotificationService(),
	)
	router := handler.NewRouter(loanService, investService)

	log.Printf("loan service listening on %s", cfg.HTTPAddress)
	if err := http.ListenAndServe(cfg.HTTPAddress, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
