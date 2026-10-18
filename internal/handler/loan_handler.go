package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"loan-service/internal/dto"
	"loan-service/internal/service"

	"github.com/gin-gonic/gin"
)

type LoanHandler struct {
	service         *service.LoanService
	investService   *service.InvestService
	disburseService *service.DisburseService
}

func NewRouter(
	loanService *service.LoanService,
	investService *service.InvestService,
	disburseService *service.DisburseService,
) *gin.Engine {
	router := gin.Default()
	handler := &LoanHandler{
		service:         loanService,
		investService:   investService,
		disburseService: disburseService,
	}

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

	loanID, err := this.service.CreateLoan(c.Request.Context(), request)
	if errors.Is(err, service.ErrBorrowerNotFound) {
		c.JSON(http.StatusNotFound, dto.MessageResponse{Message: err.Error()})
		return
	}
	if err != nil {
		log.Printf("create loan failed: %v", err)
		c.JSON(http.StatusInternalServerError, dto.MessageResponse{Message: "failed to create loan"})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateLoanResponse{ID: loanID, Message: "loan created successfully"})
}

func (this *LoanHandler) getLoan(c *gin.Context) {
	c.JSON(http.StatusOK, this.service.GetLoan(c.Param("id")))
}

func (this *LoanHandler) approveLoan(c *gin.Context) {
	loanID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || loanID == 0 {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: "invalid loan ID"})
		return
	}

	var request dto.ApproveLoanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	err = this.service.ApproveLoan(c.Request.Context(), loanID, request)
	switch {
	case errors.Is(err, service.ErrInvalidApprovalRequest):
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
	case errors.Is(err, service.ErrLoanNotFound):
		c.JSON(http.StatusNotFound, dto.MessageResponse{Message: err.Error()})
	case errors.Is(err, service.ErrLoanNotProposed), errors.Is(err, service.ErrApprovalConflict):
		c.JSON(http.StatusConflict, dto.MessageResponse{Message: err.Error()})
	case err != nil:
		log.Printf("approve loan failed: %v", err)
		c.JSON(http.StatusInternalServerError, dto.MessageResponse{Message: "failed to approve loan"})
	default:
		c.JSON(http.StatusOK, dto.MessageResponse{Message: "loan approved successfully"})
	}
}

func (this *LoanHandler) investInLoan(c *gin.Context) {
	loanID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || loanID == 0 {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: "invalid loan ID"})
		return
	}

	var request dto.InvestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	err = this.investService.Invest(c.Request.Context(), loanID, request)
	switch {
	case errors.Is(err, service.ErrInvalidInvestmentAmount):
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
	case errors.Is(err, service.ErrLoanNotFound), errors.Is(err, service.ErrInvestorNotFound):
		c.JSON(http.StatusNotFound, dto.MessageResponse{Message: err.Error()})
	case errors.Is(err, service.ErrLoanNotApprovedForInvest), errors.Is(err, service.ErrInvestmentExceedsLoan):
		c.JSON(http.StatusConflict, dto.MessageResponse{Message: err.Error()})
	case err != nil:
		log.Printf("invest in loan failed: %v", err)
		c.JSON(http.StatusInternalServerError, dto.MessageResponse{Message: "failed to invest in loan"})
	default:
		c.JSON(http.StatusOK, dto.MessageResponse{Message: "investment created successfully"})
	}
}

func (this *LoanHandler) disburseLoan(c *gin.Context) {
	loanID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || loanID == 0 {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: "invalid loan ID"})
		return
	}

	var request dto.DisburseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
		return
	}

	disbursementID, err := this.disburseService.Disburse(c.Request.Context(), loanID, request)
	switch {
	case errors.Is(err, service.ErrInvalidDisbursementRequest),
		errors.Is(err, service.ErrInvalidSignedAgreementURL):
		c.JSON(http.StatusBadRequest, dto.MessageResponse{Message: err.Error()})
	case errors.Is(err, service.ErrLoanNotFound):
		c.JSON(http.StatusNotFound, dto.MessageResponse{Message: err.Error()})
	case errors.Is(err, service.ErrLoanNotInvested),
		errors.Is(err, service.ErrDisbursementConflict):
		c.JSON(http.StatusConflict, dto.MessageResponse{Message: err.Error()})
	case err != nil:
		log.Printf("disburse loan failed: %v", err)
		c.JSON(http.StatusInternalServerError, dto.MessageResponse{Message: "failed to disburse loan"})
	default:
		c.JSON(http.StatusCreated, dto.DisbursementResponse{
			DisbursementID: disbursementID,
		})
	}
}
