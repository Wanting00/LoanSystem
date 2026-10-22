package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"loan-service/config"
	"loan-service/internal/handler"
	"loan-service/internal/repository"
	"loan-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := config.OpenMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}

	loanRepository := repository.NewLoanRepository(db)
	borrowerService := service.NewBorrowerService(repository.NewBorrowerRepository(db))
	loanService := service.NewLoanService(db, loanRepository, borrowerService)
	investmentRepository := repository.NewInvestmentRepository()
	investorRepository := repository.NewInvestorRepository()
	notificationLogRepository := repository.NewNotificationLogRepository(db)
	notificationService := service.NewNotificationService(
		notificationLogRepository,
	)
	notificationWorker := service.NewNotificationWorker(notificationLogRepository, notificationService)
	go notificationWorker.Run(context.Background())

	investService := service.NewInvestService(
		db,
		loanRepository,
		investmentRepository,
		investorRepository,
		notificationLogRepository,
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
