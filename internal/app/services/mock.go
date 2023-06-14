package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAccrualService struct {
	mock.Mock
}

func (m *MockAccrualService) CalcOrderAccrual(ctx context.Context, orderNumber string) (*CalcOrderAccrualResponse, error) {
	args := m.Called()
	return nil, args.Error(0)
}
