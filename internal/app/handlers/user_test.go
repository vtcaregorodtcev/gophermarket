package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/storage"
	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name                   string
		requestBody            interface{}
		needMockGetUserByLogin bool
		mockGetUserByLogin     *models.User
		mockGetUserByLoginErr  error
		needMockCreateUser     bool
		mockCreateUser         *models.User
		mockCreateUserErr      error
		expectedCode           int
		expectedBody           string
	}{
		{
			name:                   "Valid registration",
			requestBody:            gin.H{"login": "testuser", "password": "password"},
			needMockGetUserByLogin: true,
			mockGetUserByLogin:     nil,
			mockGetUserByLoginErr:  nil,
			needMockCreateUser:     true,
			mockCreateUser:         &models.User{ID: 1, Login: "testuser", Password: "password"},
			mockCreateUserErr:      nil,
			expectedCode:           http.StatusOK,
			expectedBody:           `{"message": "User successfully registered and authenticated"}`,
		},
		{
			name:                   "Empty credentials",
			requestBody:            gin.H{"login": "", "password": ""},
			needMockGetUserByLogin: false,
			mockGetUserByLogin:     nil,
			mockGetUserByLoginErr:  nil,
			needMockCreateUser:     false,
			mockCreateUser:         nil,
			mockCreateUserErr:      nil,
			expectedCode:           http.StatusBadRequest,
			expectedBody:           `{"error": "Empty credentials"}`,
		},
		{
			name:                   "Login already taken",
			requestBody:            gin.H{"login": "existinguser", "password": "password"},
			needMockGetUserByLogin: true,
			mockGetUserByLogin:     &models.User{ID: 1, Login: "existinguser", Password: "password"},
			mockGetUserByLoginErr:  nil,
			needMockCreateUser:     false,
			mockCreateUser:         nil,
			mockCreateUserErr:      nil,
			expectedCode:           http.StatusConflict,
			expectedBody:           `{"error": "Login is already taken"}`,
		},
		{
			name:                   "Storage error",
			requestBody:            gin.H{"login": "testuser", "password": "password"},
			needMockGetUserByLogin: true,
			mockGetUserByLogin:     nil,
			mockGetUserByLoginErr:  errors.New("Something went wrong"),
			needMockCreateUser:     false,
			mockCreateUser:         nil,
			mockCreateUserErr:      nil,
			expectedCode:           http.StatusInternalServerError,
			expectedBody:           `{"error": "Something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			storageMock := &storage.MockStorager{}
			handler := NewUserHandler(storageMock, nil, nil)
			router.POST("/api/user/register", handler.Register)

			if tt.needMockGetUserByLogin {
				storageMock.On("GetUserByLogin", mock.Anything, tt.requestBody.(gin.H)["login"]).Return(tt.mockGetUserByLogin, tt.mockGetUserByLoginErr)
			}

			if tt.needMockCreateUser {
				storageMock.On("CreateUser", mock.Anything, tt.requestBody.(gin.H)["login"], tt.requestBody.(gin.H)["password"]).Return(tt.mockCreateUser, tt.mockCreateUserErr)
			}

			jsonStr, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer(jsonStr))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			storageMock.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		needMockGetUser  bool
		mockGetUser      *models.User
		mockGetUserErr   error
		expectedCode     int
		expectedBody     string
		isExpectedCookie bool
	}{
		{
			name:            "Valid login",
			requestBody:     gin.H{"login": "testuser", "password": "password"},
			needMockGetUser: true,
			mockGetUser: &models.User{
				ID:       1,
				Login:    "testuser",
				Password: "password",
			},
			mockGetUserErr:   nil,
			expectedCode:     http.StatusOK,
			expectedBody:     `{"message": "Authentication successful"}`,
			isExpectedCookie: true,
		},
		{
			name:             "Empty credentials",
			requestBody:      gin.H{"login": "", "password": ""},
			needMockGetUser:  false,
			mockGetUser:      nil,
			mockGetUserErr:   nil,
			expectedCode:     http.StatusBadRequest,
			expectedBody:     `{"error": "Empty credentials"}`,
			isExpectedCookie: false,
		},
		{
			name:             "Invalid credentials",
			requestBody:      gin.H{"login": "testuser", "password": "password"},
			needMockGetUser:  true,
			mockGetUser:      nil,
			mockGetUserErr:   nil,
			expectedCode:     http.StatusUnauthorized,
			expectedBody:     `{"error": "Invalid credentials"}`,
			isExpectedCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			storageMock := &storage.MockStorager{}
			handler := NewUserHandler(storageMock, nil, nil)
			router.POST("/api/user/login", handler.Login)

			if tt.needMockGetUser {
				storageMock.On("GetUserByLogin", mock.Anything, tt.requestBody.(gin.H)["login"]).Return(tt.mockGetUser, tt.mockGetUserErr)
			}

			jsonStr, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(jsonStr))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			if tt.isExpectedCookie {
				assert.Equal(t, len(w.Header().Get("Set-Cookie")) > 0, true)
			}

			storageMock.AssertExpectations(t)
		})
	}
}

type mockWorkerPool struct {
	mock.Mock
}

func (m *mockWorkerPool) Submit(f func()) {}

func TestUserHandler_SubmitOrder(t *testing.T) {
	tests := []struct {
		name        string
		requestBody interface{}
		userID      float64

		needMockGetOrderByNumber bool
		mockGetOrderByNumber     *models.Order
		mockGetOrderErr          error

		needMockUpdateOrderAccrualAndUserBalance bool
		needMockCalcAccrual                      bool
		mockCalcAccrual                          *services.CalcOrderAccrualResponse

		needMockCreateOrder bool
		mockCreateOrder     *models.Order
		mockCreateOrderErr  error

		expectedCode int
		expectedBody string
	}{
		{
			name:                                     "Valid order submission",
			requestBody:                              "123456789106",
			userID:                                   1,
			needMockGetOrderByNumber:                 true,
			needMockCreateOrder:                      true,
			needMockCalcAccrual:                      true,
			needMockUpdateOrderAccrualAndUserBalance: true,
			mockCalcAccrual: &services.CalcOrderAccrualResponse{
				Order:   "123456789106",
				Accrual: 100.0,
				Status:  models.PROCESSED,
			},
			mockGetOrderByNumber: nil,
			mockGetOrderErr:      nil,
			mockCreateOrder: &models.Order{
				ID:     1,
				Number: "123456789106",
				UserID: 1,
				Status: models.NEW,
			},
			mockCreateOrderErr: nil,
			expectedCode:       http.StatusAccepted,
			expectedBody:       `{"id":1,"number":"123456789106","status":"NEW","user_id":1,"uploaded_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:                     "Empty request body",
			requestBody:              "",
			userID:                   1,
			needMockGetOrderByNumber: false,
			mockGetOrderByNumber:     nil,
			mockGetOrderErr:          nil,
			needMockCreateOrder:      false,
			mockCreateOrder:          nil,
			mockCreateOrderErr:       nil,
			expectedCode:             http.StatusBadRequest,
			expectedBody:             `{"error": "Empty request body"}`,
		},
		{
			name:                 "Invalid order format",
			requestBody:          "abcde",
			userID:               1,
			mockGetOrderByNumber: nil,
			mockGetOrderErr:      nil,
			mockCreateOrder:      nil,
			mockCreateOrderErr:   nil,
			expectedCode:         http.StatusUnprocessableEntity,
			expectedBody:         `{"error": "Wrong order format"}`,
		},
		{
			name:                     "Existing order",
			requestBody:              "123456789106",
			userID:                   1,
			needMockGetOrderByNumber: true,
			mockGetOrderByNumber: &models.Order{
				ID:     1,
				Number: "123456789106",
				UserID: 2,
			},
			mockGetOrderErr:    nil,
			mockCreateOrder:    nil,
			mockCreateOrderErr: nil,
			expectedCode:       http.StatusConflict,
			expectedBody:       `{"error": "Order is already submitted"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			router.Use(func(c *gin.Context) {
				c.Set("userID", tt.userID)
			})

			storageMock := &storage.MockStorager{}
			accrualServiceMock := &services.MockAccrualService{}
			workerPoolMock := &mockWorkerPool{}
			handler := NewUserHandler(storageMock, accrualServiceMock, workerPoolMock)
			router.POST("/api/user/orders", handler.SubmitOrder)

			if tt.needMockGetOrderByNumber {
				storageMock.On("GetOrderByNumber", mock.Anything, mock.Anything).Return(tt.mockGetOrderByNumber, tt.mockGetOrderErr)
			}

			if tt.needMockCreateOrder {
				storageMock.On("CreateOrder", mock.Anything, tt.requestBody, uint(tt.userID)).Return(tt.mockCreateOrder, tt.mockCreateOrderErr)
			}

			req, _ := http.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(tt.requestBody.(string)))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			storageMock.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetOrders(t *testing.T) {
	tests := []struct {
		name             string
		userID           float64
		mockGetOrders    *[]*models.Order
		mockGetOrdersErr error
		expectedCode     int
		expectedBody     string
	}{
		{
			name:             "Successful retrieval of orders",
			userID:           1,
			mockGetOrders:    &[]*models.Order{{ID: 1, Number: "123456789106"}, {ID: 2, Number: "67890"}},
			mockGetOrdersErr: nil,
			expectedCode:     http.StatusOK,
			expectedBody:     `[{"id":1,"number":"123456789106","status":"","uploaded_at":"0001-01-01T00:00:00Z"},{"id":2,"number":"67890","status":"","uploaded_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:             "No orders found",
			userID:           1,
			mockGetOrders:    &[]*models.Order{},
			mockGetOrdersErr: nil,
			expectedCode:     http.StatusNoContent,
			expectedBody:     "",
		},
		{
			name:             "Error retrieving orders",
			userID:           1,
			mockGetOrders:    nil,
			mockGetOrdersErr: errors.New("Something went wrong"),
			expectedCode:     http.StatusInternalServerError,
			expectedBody:     `{"error":"Could not retrieve orders"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			router.Use(func(c *gin.Context) {
				c.Set("userID", tt.userID)
			})
			storageMock := &storage.MockStorager{}
			handler := NewUserHandler(storageMock, nil, nil)
			router.GET("/api/user/orders", handler.GetOrders)

			storageMock.On("GetOrdersByUserID", uint(tt.userID)).Return(tt.mockGetOrders, tt.mockGetOrdersErr)

			req, _ := http.NewRequest("GET", "/api/user/orders", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			body := w.Body.String()

			if body != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			storageMock.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetBalance(t *testing.T) {
	tests := []struct {
		name               string
		userID             float64
		mockGetUserByID    *models.User
		mockGetUserByIDErr error
		expectedCode       int
		expectedBody       string
	}{
		{
			name:               "Successful retrieval of balance",
			userID:             1,
			mockGetUserByID:    &models.User{ID: 1, Balance: 100.0, Withdrawn: 50.0},
			mockGetUserByIDErr: nil,
			expectedCode:       http.StatusOK,
			expectedBody:       `{"current":100,"withdrawn":50}`,
		},
		{
			name:               "Error retrieving user",
			userID:             1,
			mockGetUserByID:    nil,
			mockGetUserByIDErr: errors.New("Something went wrong"),
			expectedCode:       http.StatusInternalServerError,
			expectedBody:       `{"error":"Could not retrieve User"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			router.Use(func(c *gin.Context) {
				c.Set("userID", tt.userID)
			})
			storageMock := &storage.MockStorager{}
			handler := NewUserHandler(storageMock, nil, nil)
			router.GET("/api/user/balance", handler.GetBalance)

			storageMock.On("GetUserByID", mock.Anything, uint(tt.userID)).Return(tt.mockGetUserByID, tt.mockGetUserByIDErr)

			req, _ := http.NewRequest("GET", "/api/user/balance", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			storageMock.AssertExpectations(t)
		})
	}
}

func TestUserHandler_WithdrawBalance(t *testing.T) {
	tests := []struct {
		name                   string
		userID                 float64
		requestBody            interface{}
		mockWithdrawBalanceErr error
		expectedCode           int
		expectedBody           string
	}{
		{
			name:                   "Successful withdrawal",
			userID:                 1,
			requestBody:            withdrawRequest{Order: "123456789", Sum: 50.0},
			mockWithdrawBalanceErr: nil,
			expectedCode:           http.StatusOK,
			expectedBody:           `{"message":"Balance withdrawal successful"}`,
		},
		{
			name:                   "Insufficient balance",
			userID:                 1,
			requestBody:            withdrawRequest{Order: "123456789", Sum: 50.0},
			mockWithdrawBalanceErr: storage.ErrInsufficientBalance,
			expectedCode:           http.StatusPaymentRequired,
			expectedBody:           `{"error":"Insufficient balance"}`,
		},
		{
			name:                   "Cannot proceed withdrawal with existing order",
			userID:                 1,
			requestBody:            withdrawRequest{Order: "123456789", Sum: 50.0},
			mockWithdrawBalanceErr: storage.ErrOrderAlreadyExists,
			expectedCode:           http.StatusPaymentRequired,
			expectedBody:           `{"error":"Cannot proceed withdrawal with existing order"}`,
		},
		{
			name:                   "Internal server error",
			userID:                 1,
			requestBody:            withdrawRequest{Order: "123456789", Sum: 50.0},
			mockWithdrawBalanceErr: errors.New("Internal server error"),
			expectedCode:           http.StatusInternalServerError,
			expectedBody:           `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			storageMock := &storage.MockStorager{}
			handler := NewUserHandler(storageMock, nil, nil)

			router.Use(func(c *gin.Context) {
				c.Set("userID", tt.userID)
			})

			router.POST("/api/user/balance/withdraw", handler.WithdrawBalance)

			storageMock.On("WithdrawBalance", mock.Anything, uint(tt.userID), tt.requestBody.(withdrawRequest).Order, tt.requestBody.(withdrawRequest).Sum).Return(tt.mockWithdrawBalanceErr)

			jsonStr, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(jsonStr))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			storageMock.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetWithdrawals(t *testing.T) {
	tests := []struct {
		name                      string
		userID                    float64
		mockGetUserWithdrawals    *[]*models.Withdrawal
		mockGetUserWithdrawalsErr error
		expectedCode              int
		expectedBody              string
	}{
		{
			name:                      "Successful retrieval",
			userID:                    1,
			mockGetUserWithdrawals:    &[]*models.Withdrawal{{OrderNumber: "123456789", Sum: 50.0}, {OrderNumber: "123456783", Sum: 250.0}},
			mockGetUserWithdrawalsErr: nil,
			expectedCode:              http.StatusOK,
			expectedBody:              `[{"order":"123456789","sum":50,"processed_at":"0001-01-01T00:00:00Z"},{"order":"123456783","sum":250,"processed_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:                      "Storage error",
			userID:                    1,
			mockGetUserWithdrawals:    nil,
			mockGetUserWithdrawalsErr: errors.New("Something went wrong"),
			expectedCode:              http.StatusInternalServerError,
			expectedBody:              `{"error":"Something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			storageMock := &storage.MockStorager{}
			handler := NewUserHandler(storageMock, nil, nil)

			router.Use(func(c *gin.Context) {
				c.Set("userID", tt.userID)
			})

			router.GET("/api/user/withdrawals", handler.GetWithdrawals)

			storageMock.On("GetUserWithdrawals", uint(tt.userID)).Return(tt.mockGetUserWithdrawals, tt.mockGetUserWithdrawalsErr)

			req, _ := http.NewRequest("GET", "/api/user/withdrawals", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			storageMock.AssertExpectations(t)
		})
	}
}
