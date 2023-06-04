package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/app/storage"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/helpers"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"
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

	existingUser, err := uh.storage.GetUserByLogin(newUser.Login)
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

	user, err := uh.storage.GetUserByLogin(credentials.Login)
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

// SubmitOrder endpoint
func (uh *UserHandler) SubmitOrder(c *gin.Context) {
	// Logic to submit an order
}

func (uh *UserHandler) GetOrders(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	orders, err := uh.storage.GetOrdersByUserId(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve orders"})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"orders": []models.Order{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (uh *UserHandler) GetBalance(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	user, err := uh.storage.GetUserByID(uint(userID))
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

// WithdrawBalance endpoint
func (uh *UserHandler) WithdrawBalance(c *gin.Context) {
	// Logic to withdraw balance
}

// GetWithdrawals endpoint
func (uh *UserHandler) GetWithdrawals(c *gin.Context) {
	// Logic to get withdrawal information
}
