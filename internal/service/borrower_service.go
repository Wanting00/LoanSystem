package service

import (
	"context"
	"errors"

	"loan-service/internal/repository"
)

var ErrBorrowerNotFound = errors.New("borrower not found")

type BorrowerService struct {
	borrowerRepository *repository.BorrowerRepository
}

func NewBorrowerService(borrowerRepository *repository.BorrowerRepository) *BorrowerService {
	return &BorrowerService{borrowerRepository: borrowerRepository}
}

func (s *BorrowerService) BorrowerExists(ctx context.Context, borrowerID uint64) (bool, error) {
	return s.borrowerRepository.ExistsByID(ctx, borrowerID)
}
