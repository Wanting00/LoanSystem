package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"loan-service/internal/dto"
	"loan-service/internal/service"
)

type LoanHandler struct {
	service *service.LoanService
}

func NewRouter(loanService *service.LoanService) *gin.Engine {
	router := gin.Default()
	handler := &LoanHandler{service: loanService}

	router.GET("/health", handler.health)

	api := router.Group("/api/v1")
	api.POST("/loans", handler.createLoan)
	api.GET("/loans/:id", handler.getLoan)
	api.POST("/loans/:id/approve", handler.approveLoan)
	api.POST("/loans/:id/invest", handler.investInLoan)
	api.POST("/loans/:id/disburse", handler.disburseLoan)

	return router
}

func (h *LoanHandler) health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

func (h *LoanHandler) createLoan(c *gin.Context) {
	var request dto.CreateLoanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, h.service.CreateLoan(request))
}

func (h *LoanHandler) getLoan(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetLoan(c.Param("id")))
}

func (h *LoanHandler) approveLoan(c *gin.Context) {
	var request dto.ApproveLoanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.service.ApproveLoan(c.Param("id"), request))
}

func (h *LoanHandler) investInLoan(c *gin.Context) {
	var request dto.InvestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.service.InvestInLoan(c.Param("id"), request))
}

func (h *LoanHandler) disburseLoan(c *gin.Context) {
	var request dto.DisburseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.service.DisburseLoan(c.Param("id"), request))
}
