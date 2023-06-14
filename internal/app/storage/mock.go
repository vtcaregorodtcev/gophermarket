package storage

import (
	"context"
	"database/sql"

	"github.com/stretchr/testify/mock"

	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

type MockStorager struct {
	mock.Mock
}

func (m *MockStorager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorager) GetUserByID(tx *sql.Tx, id uint) (*models.User, error) {
	args := m.Called(tx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorager) GetUserByLogin(tx *sql.Tx, login string) (*models.User, error) {
	args := m.Called(tx, login)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorager) CreateUser(ctx context.Context, login, password string) (*models.User, error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorager) GetOrderByNumber(tx *sql.Tx, orderNumber string) (*models.Order, error) {
	args := m.Called(tx, orderNumber)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockStorager) CreateOrder(ctx context.Context, orderNumber string, userID uint) (*models.Order, error) {
	args := m.Called(ctx, orderNumber, userID)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockStorager) UpdateOrderAccrualAndUserBalance(ctx context.Context, orderID uint, userID uint, accrualResp *services.CalcOrderAccrualResponse) error {
	args := m.Called(ctx, orderID, userID, accrualResp)
	return args.Error(0)
}

func (m *MockStorager) GetOrdersByUserID(userID uint) (*[](*models.Order), error) {
	args := m.Called(userID)
	return args.Get(0).(*[](*models.Order)), args.Error(1)
}

func (m *MockStorager) WithdrawBalance(ctx context.Context, userID uint, orderNumber string, withdrawalAmount float64) error {
	args := m.Called(ctx, userID, orderNumber, withdrawalAmount)
	return args.Error(0)
}

func (m *MockStorager) GetUserWithdrawals(userID uint) (*[](*models.Withdrawal), error) {
	args := m.Called(userID)
	return args.Get(0).(*[](*models.Withdrawal)), args.Error(1)
}
