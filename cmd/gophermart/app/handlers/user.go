package handlers

import (
	"net/http"
	"time"

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
}

func (uh *UserHandler) Login(c *gin.Context) {
	var credentials models.User
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uh.storage.GetUserByUsername(credentials.Username)
	if err != nil || user == nil || user.Password != credentials.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	hash, err := helpers.CredentialsHash(user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.SetCookie("user", hash, int(time.Now().Add(24*time.Hour).Unix()), "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{"message": "Authentication successful"})
}

// SubmitOrder endpoint
func (uh *UserHandler) SubmitOrder(c *gin.Context) {
	// Logic to submit an order
}

// GetOrders endpoint
func (uh *UserHandler) GetOrders(c *gin.Context) {
	// Logic to get a list of orders submitted by a user
}

// GetBalance endpoint
func (uh *UserHandler) GetBalance(c *gin.Context) {
	// Logic to get the current balance of a user
}

// WithdrawBalance endpoint
func (uh *UserHandler) WithdrawBalance(c *gin.Context) {
	// Logic to withdraw balance
}

// GetWithdrawals endpoint
func (uh *UserHandler) GetWithdrawals(c *gin.Context) {
	// Logic to get withdrawal information
}
