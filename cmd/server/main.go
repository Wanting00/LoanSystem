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
	investmentRepository := repository.NewInvestmentRepository()
	investorRepository := repository.NewInvestorRepository()
	notificationService := service.NewNotificationService(
		db,
		investmentRepository,
		investorRepository,
		repository.NewNotificationLogRepository(db),
	)
	investService := service.NewInvestService(
		db,
		loanRepository,
		investmentRepository,
		investorRepository,
		notificationService,
	)
	disburseService := service.NewDisburseService(
		db,
		loanRepository,
		repository.NewDisbursementRepository(),
	)
	router := handler.NewRouter(loanService, investService, disburseService)

	log.Printf("loan service listening on %s", cfg.HTTPAddress)
	if err := http.ListenAndServe(cfg.HTTPAddress, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
