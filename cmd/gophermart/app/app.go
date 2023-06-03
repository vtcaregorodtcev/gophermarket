package app

import (
	"fmt"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/app/handlers"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/app/middleware"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/app/storage"

	"github.com/gin-gonic/gin"
)

type App struct {
	router  *gin.Engine
	storage *storage.Storage
}

func New() *App {
	router := gin.Default()
	storage := storage.New()

	app := &App{
		router:  router,
		storage: storage,
	}

	return app
}

func (app *App) Run() {
	userHandler := handlers.NewUserHandler(app.storage)

	userApi := app.router.Group("/api/user")
	{
		userApi.POST("/register", userHandler.Register)
		userApi.POST("/login", userHandler.Login)

		userApi.Use(middleware.Auth())
		{
			userApi.POST("/orders", userHandler.SubmitOrder)
			userApi.GET("/orders", userHandler.GetOrders)
			userApi.GET("/balance", userHandler.GetBalance)
			userApi.POST("/balance/withdraw", userHandler.WithdrawBalance)
			userApi.GET("/withdrawals", userHandler.GetWithdrawals)
		}
	}

	app.router.Run(":8080")
}

func (app *App) Shutdown() {
	err := app.storage.Close()
	if err != nil {
		fmt.Println("db close error:", err)
	}
}
