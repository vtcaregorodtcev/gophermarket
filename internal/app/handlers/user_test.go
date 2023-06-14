package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gammazero/workerpool"
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
			workerPoolMock := workerpool.New(0)
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
			accrualServiceMock.AssertExpectations(t)
		})
	}
}
