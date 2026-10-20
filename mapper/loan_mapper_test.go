package mapper

import (
	"testing"
	"time"

	"loan-service/internal/dto"
	"loan-service/internal/model"

	"github.com/shopspring/decimal"
)

func TestCreateLoanRequestToLoan(t *testing.T) {
	principal := decimal.RequireFromString("25000")
	rate := decimal.RequireFromString("0.05")
	roi := decimal.RequireFromString("0.08")
	request := dto.CreateLoanRequest{
		BorrowerID:         42,
		Principal:          &principal,
		Rate:               &rate,
		ROI:                &roi,
		AgreementLetterURL: "https://example.com/agreement.pdf",
	}

	got := CreateLoanRequestToLoan(request)

	if got.BorrowerID != request.BorrowerID {
		t.Errorf("BorrowerID = %d, want %d", got.BorrowerID, request.BorrowerID)
	}
	if !got.Principal.Equal(principal) {
		t.Errorf("Principal = %s, want %s", got.Principal, principal)
	}
	if !got.Rate.Equal(rate) {
		t.Errorf("Rate = %s, want %s", got.Rate, rate)
	}
	if !got.ROI.Equal(roi) {
		t.Errorf("ROI = %s, want %s", got.ROI, roi)
	}
	if got.AgreementLetterURL != request.AgreementLetterURL {
		t.Errorf("AgreementLetterURL = %q, want %q", got.AgreementLetterURL, request.AgreementLetterURL)
	}
	if got.Status != model.LoanStatusProposed {
		t.Errorf("Status = %q, want %q", got.Status, model.LoanStatusProposed)
	}
}

func TestApproveLoanRequestToApproval(t *testing.T) {
	approvalDate := time.Date(2025, time.January, 15, 10, 30, 0, 0, time.UTC)
	request := dto.ApproveLoanRequest{
		ValidatorPhotoURL: "https://example.com/validator.jpg",
		EmployeeID:        "employee-7",
		ApprovalDate:      approvalDate,
	}

	got := ApproveLoanRequestToApproval(123, request)

	if got.LoanID != 123 {
		t.Errorf("LoanID = %d, want 123", got.LoanID)
	}
	if got.ValidatorPhotoURL != request.ValidatorPhotoURL {
		t.Errorf("ValidatorPhotoURL = %q, want %q", got.ValidatorPhotoURL, request.ValidatorPhotoURL)
	}
	if got.EmployeeID != request.EmployeeID {
		t.Errorf("EmployeeID = %q, want %q", got.EmployeeID, request.EmployeeID)
	}
	if !got.ApprovalDate.Equal(approvalDate) {
		t.Errorf("ApprovalDate = %v, want %v", got.ApprovalDate, approvalDate)
	}
}

func TestLoanToLoanResponse(t *testing.T) {
	loan := model.Loan{
		ID:                 123,
		BorrowerID:         42,
		Principal:          decimal.RequireFromString("25000.00"),
		Rate:               decimal.RequireFromString("0.050000"),
		ROI:                decimal.RequireFromString("0.080000"),
		Status:             model.LoanStatusApproved,
		AgreementLetterURL: "https://example.com/agreement.pdf",
	}

	got := LoanToLoanResponse(loan)

	if got.ID != loan.ID {
		t.Errorf("ID = %d, want %d", got.ID, loan.ID)
	}
	if got.BorrowerID != loan.BorrowerID {
		t.Errorf("BorrowerID = %d, want %d", got.BorrowerID, loan.BorrowerID)
	}
	if !got.Principal.Equal(loan.Principal) {
		t.Errorf("Principal = %s, want %s", got.Principal, loan.Principal)
	}
	if !got.Rate.Equal(loan.Rate) {
		t.Errorf("Rate = %s, want %s", got.Rate, loan.Rate)
	}
	if !got.ROI.Equal(loan.ROI) {
		t.Errorf("ROI = %s, want %s", got.ROI, loan.ROI)
	}
	if got.Status != loan.Status {
		t.Errorf("Status = %q, want %q", got.Status, loan.Status)
	}
	if got.AgreementLetterURL != loan.AgreementLetterURL {
		t.Errorf("AgreementLetterURL = %q, want %q", got.AgreementLetterURL, loan.AgreementLetterURL)
	}
}
