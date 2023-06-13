package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/storage"
	"github.com/vtcaregorodtcev/gophermarket/internal/helpers"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

type UserHandler struct {
	storage *storage.Storage
}

func NewUserHandler(storage *storage.Storage) *UserHandler {
	return &UserHandler{storage: storage}
}

func (uh *UserHandler) Register(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(newUser.Login) == 0 || len(newUser.Password) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty credentials"})
		return
	}

	existingUser, err := uh.storage.GetUserByLogin(nil, newUser.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Login is already taken"})
		return
	}

	user, err := uh.storage.CreateUser(c.Request.Context(), newUser.Login, newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	jwt, err := helpers.GetJWTByID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.SetCookie("jwt", jwt, 24*60*60, "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{"message": "User successfully registered and authenticated"})
}

func (uh *UserHandler) Login(c *gin.Context) {
	var credentials models.User
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(credentials.Login) == 0 || len(credentials.Password) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty credentials"})
		return
	}

	user, err := uh.storage.GetUserByLogin(nil, credentials.Login)
	if err != nil || user == nil || user.Password != credentials.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	jwt, err := helpers.GetJWTByID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.SetCookie("jwt", jwt, 24*60*60, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Authentication successful"})
}

func (uh *UserHandler) SubmitOrder(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	orderNumber := string(body)
	logger.Infof("SubmitOrder with order number: %s", orderNumber)

	if len(orderNumber) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty request body"})
		return
	}

	if isValid := helpers.IsOrderNumberValid(orderNumber); !isValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Wrong order format"})
		return
	}

	existingOrder, err := uh.storage.GetOrderByNumber(nil, orderNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	if existingOrder != nil {
		if existingOrder.UserID != uint(userID) {
			c.JSON(http.StatusConflict, gin.H{"error": "Order is already submitted"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Order is already submitted by this user"})
		}
		return
	}

	order, err := uh.storage.CreateOrder(c.Request.Context(), orderNumber, uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	go func() {
		accrualService := services.GetAccrualServiceInstance()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		resp, err := accrualService.CalcOrderAccrual(ctx, order.Number)
		if err != nil {
			logger.Infof("submit order: CalcOrderAccrual: %v", err)
			return
		}

		err = uh.storage.UpdateOrderAccrualAndUserBalance(ctx, order.ID, uint(userID), resp)
		if err != nil {
			logger.Infof("submit order: UpdateOrderAccrualAndUserBalance: %v", err)
			return
		}
	}()

	c.JSON(http.StatusAccepted, order)
}

func (uh *UserHandler) GetOrders(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	orders, err := uh.storage.GetOrdersByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve orders"})
		return
	}

	if len(*orders) == 0 {
		c.JSON(http.StatusNoContent, []models.Order{})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (uh *UserHandler) GetBalance(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	user, err := uh.storage.GetUserByID(nil, uint(userID))
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve User"})
		return
	}

	type data struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}

	c.JSON(http.StatusOK, data{Current: user.Balance, Withdrawn: user.Withdrawn})
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (uh *UserHandler) WithdrawBalance(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	var req withdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err := uh.storage.WithdrawBalance(c.Request.Context(), uint(userID), req.Order, req.Sum)
	if err != nil {
		switch err {
		case storage.ErrInsufficientBalance:
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
		case storage.ErrOrderAlreadyExists:
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Cannot proceed withdrawal with existing order"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Balance withdrawal successful"})
	}
}

func (uh *UserHandler) GetWithdrawals(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	withdrawals, err := uh.storage.GetUserWithdrawals(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	if len(*withdrawals) == 0 {
		c.JSON(http.StatusNoContent, []models.Order{})
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}
