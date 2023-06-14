package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAccrualService struct {
	mock.Mock
}

func (m *MockAccrualService) CalcOrderAccrual(ctx context.Context, orderNumber string) (*CalcOrderAccrualResponse, error) {
	args := m.Called(ctx, orderNumber)

	response := args.Get(0)
	err := args.Error(1)

	if response != nil {
		return response.(*CalcOrderAccrualResponse), err
	}
	return nil, err
}
