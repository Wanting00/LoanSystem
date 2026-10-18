package mapper

import (
	"loan-service/internal/dto"
	"loan-service/internal/model"
)

func CreateLoanRequestToLoan(request dto.CreateLoanRequest) model.Loan {
	return model.Loan{
		BorrowerID:         request.BorrowerID,
		Principal:          request.Principal,
		Rate:               request.Rate,
		ROI:                request.ROI,
		AgreementLetterURL: request.AgreementLetterURL,
		Status:             model.LoanStatusProposed,
	}
}

func ApproveLoanRequestToApproval(loanID uint64, request dto.ApproveLoanRequest) model.Approval {
	return model.Approval{
		LoanID:            loanID,
		ValidatorPhotoURL: request.ValidatorPhotoURL,
		EmployeeID:        request.EmployeeID,
		ApprovalDate:      request.ApprovalDate,
	}
}

func LoanToLoanResponse(loan model.Loan) dto.LoanResponse {
	return dto.LoanResponse{
		ID:                 loan.ID,
		BorrowerID:         loan.BorrowerID,
		Principal:          loan.Principal,
		Rate:               loan.Rate,
		ROI:                loan.ROI,
		Status:             loan.Status,
		AgreementLetterURL: loan.AgreementLetterURL,
	}
}
