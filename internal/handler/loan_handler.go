package handler

import (
	"errors"
	"log"
	"net/http"

	"loan-service/internal/dto"
	"loan-service/internal/service"

	"github.com/gin-gonic/gin"
)

type LoanHandler struct {
	service *service.LoanService
}

func NewRouter(loanService *service.LoanService) *gin.Engine {
	router := gin.Default()
	handler := &LoanHandler{service: loanService}

	router.GET("/health", handler.health)

	api := router.Group("/api/v1")
	api.POST("/loans/create", handler.createLoan)
	api.GET("/loans/:id", handler.getLoan)
	api.POST("/loans/:id/approve", handler.approveLoan)
	api.POST("/loans/:id/invest", handler.investInLoan)
	api.POST("/loans/:id/disburse", handler.disburseLoan)

	return router
}

func (this *LoanHandler) health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

func (this *LoanHandler) createLoan(c *gin.Context) {
	var request dto.CreateLoanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	err := this.service.CreateLoan(c.Request.Context(), request)
	if errors.Is(err, service.ErrBorrowerNotFound) {
		c.JSON(http.StatusNotFound, dto.MessageResponse{Message: err.Error()})
		return
	}
	if err != nil {
		log.Printf("create loan failed: %v", err)
		c.JSON(http.StatusInternalServerError, dto.MessageResponse{Message: "failed to create loan"})
		return
	}

	c.JSON(http.StatusCreated, dto.MessageResponse{Message: "loan created successfully"})
}

func (this *LoanHandler) getLoan(c *gin.Context) {
	c.JSON(http.StatusOK, this.service.GetLoan(c.Param("id")))
}

func (this *LoanHandler) approveLoan(c *gin.Context) {
	var request dto.ApproveLoanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, this.service.ApproveLoan(c.Param("id"), request))
}

func (this *LoanHandler) investInLoan(c *gin.Context) {
	var request dto.InvestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, this.service.InvestInLoan(c.Param("id"), request))
}

func (this *LoanHandler) disburseLoan(c *gin.Context) {
	var request dto.DisburseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, this.service.DisburseLoan(c.Param("id"), request))
}
