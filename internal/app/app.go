package app

import (
	"github.com/vtcaregorodtcev/gophermarket/internal/app/handlers"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/middleware"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/storage"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/storage/connector/postgres"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"

	"github.com/gin-gonic/gin"
)

type Config struct {
	AccrualAddr string
	DatabaseURI string
	Addr        string
}

type App struct {
	cfg     Config
	router  *gin.Engine
	storage *storage.Storage
}

func New(cfg Config) *App {
	router := gin.Default()

	storage, err := storage.New(cfg.DatabaseURI, &postgres.PGConnector{})
	if err != nil {
		logger.Infof("storage creating error: %v", err)
	}

	services.NewAccrualService(cfg.AccrualAddr)

	app := &App{
		cfg:     cfg,
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

	err := app.router.Run(app.cfg.Addr)
	if err != nil {
		logger.Infof("app starting err: %v", err)
	}
}

func (app *App) Shutdown() {
	err := app.storage.Close()
	if err != nil {
		logger.Infof("db close error: %v", err)
	}
}
