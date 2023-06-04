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

	userAPI := app.router.Group("/api/user")
	{
		userAPI.POST("/register", userHandler.Register)
		userAPI.POST("/login", userHandler.Login)

		userAPI.Use(middleware.Auth())
		{
			userAPI.POST("/orders", userHandler.SubmitOrder)
			userAPI.GET("/orders", userHandler.GetOrders)
			userAPI.GET("/balance", userHandler.GetBalance)
			userAPI.POST("/balance/withdraw", userHandler.WithdrawBalance)
			userAPI.GET("/withdrawals", userHandler.GetWithdrawals)
		}
	}

	err := app.router.Run(":8080")
	if err != nil {
		fmt.Println("app starting err:", err)
	}
}

func (app *App) Shutdown() {
	err := app.storage.Close()
	if err != nil {
		fmt.Println("db close error:", err)
	}
}
