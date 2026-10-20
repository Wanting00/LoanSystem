package service

import (
	"errors"
	"strings"
	"testing"

	"loan-service/internal/dto"
	"loan-service/internal/model"

	"github.com/shopspring/decimal"
)

func TestIsValidIdempotencyKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{name: "empty", key: "", want: false},
		{name: "printable ASCII", key: "loan-2025-001", want: true},
		{name: "minimum printable ASCII", key: "!", want: true},
		{name: "maximum printable ASCII", key: "~", want: true},
		{name: "255 characters", key: strings.Repeat("a", 255), want: true},
		{name: "256 characters", key: strings.Repeat("a", 256), want: false},
		{name: "space", key: "loan key", want: false},
		{name: "control character", key: "loan\tkey", want: false},
		{name: "non-ASCII", key: "loan-é", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidIdempotencyKey(test.key); got != test.want {
				t.Errorf("IsValidIdempotencyKey(%q) = %t, want %t", test.key, got, test.want)
			}
		})
	}
}

func TestMatchingLoanID(t *testing.T) {
	request := dto.CreateLoanRequest{
		BorrowerID:         42,
		Principal:          decimalPtr("25000.00"),
		Rate:               decimalPtr("0.0500"),
		ROI:                decimalPtr("0.0800"),
		AgreementLetterURL: "https://example.com/agreement.pdf",
	}
	loan := model.Loan{
		ID:                 123,
		BorrowerID:         request.BorrowerID,
		Principal:          decimal.RequireFromString("25000"),
		Rate:               decimal.RequireFromString("0.05"),
		ROI:                decimal.RequireFromString("0.08"),
		AgreementLetterURL: request.AgreementLetterURL,
	}

	t.Run("matching values with equivalent decimal scales", func(t *testing.T) {
		got, err := matchingLoanID(loan, request)
		if err != nil {
			t.Fatalf("matchingLoanID() error = %v, want nil", err)
		}
		if got != loan.ID {
			t.Errorf("matchingLoanID() = %d, want %d", got, loan.ID)
		}
	})

	tests := []struct {
		name   string
		mutate func(*model.Loan)
	}{
		{
			name: "different borrower",
			mutate: func(loan *model.Loan) {
				loan.BorrowerID++
			},
		},
		{
			name: "different principal",
			mutate: func(loan *model.Loan) {
				loan.Principal = loan.Principal.Add(decimal.NewFromInt(1))
			},
		},
		{
			name: "different rate",
			mutate: func(loan *model.Loan) {
				loan.Rate = loan.Rate.Add(decimal.NewFromFloat(0.01))
			},
		},
		{
			name: "different ROI",
			mutate: func(loan *model.Loan) {
				loan.ROI = loan.ROI.Add(decimal.NewFromFloat(0.01))
			},
		},
		{
			name: "different agreement URL",
			mutate: func(loan *model.Loan) {
				loan.AgreementLetterURL += "?revision=2"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mismatchedLoan := loan
			test.mutate(&mismatchedLoan)

			got, err := matchingLoanID(mismatchedLoan, request)
			if !errors.Is(err, ErrIdempotencyConflict) {
				t.Fatalf("matchingLoanID() error = %v, want %v", err, ErrIdempotencyConflict)
			}
			if got != 0 {
				t.Errorf("matchingLoanID() = %d, want 0 on conflict", got)
			}
		})
	}
}

func TestReplayedLoan(t *testing.T) {
	request := dto.CreateLoanRequest{
		BorrowerID:         42,
		Principal:          decimalPtr("25000"),
		Rate:               decimalPtr("0.05"),
		ROI:                decimalPtr("0.08"),
		AgreementLetterURL: "https://example.com/agreement.pdf",
	}
	loan := model.Loan{
		ID:                 123,
		BorrowerID:         request.BorrowerID,
		Principal:          *request.Principal,
		Rate:               *request.Rate,
		ROI:                *request.ROI,
		AgreementLetterURL: request.AgreementLetterURL,
	}

	got, err := replayedLoan(loan, request)
	if !errors.Is(err, ErrLoanAlreadyCreated) {
		t.Fatalf("replayedLoan() error = %v, want %v", err, ErrLoanAlreadyCreated)
	}
	if got != loan.ID {
		t.Errorf("replayedLoan() = %d, want %d", got, loan.ID)
	}
}

func decimalPtr(value string) *decimal.Decimal {
	result := decimal.RequireFromString(value)
	return &result
}
