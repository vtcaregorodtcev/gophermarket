package app

import (
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"

	"github.com/vtcaregorodtcev/gophermarket/internal/app/handlers"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/middleware"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/storage"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
)

type Config struct {
	AccrualAddr string
	DatabaseURI string
	Addr        string
	PoolCount   int
}

type App struct {
	cfg            Config
	router         *gin.Engine
	storage        storage.Storager
	accrualService *services.AccrualService
	pool           *workerpool.WorkerPool
}

func New(cfg Config) (*App, error) {
	router := gin.Default()
	storage, err := storage.New(cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}

	wp := workerpool.New(cfg.PoolCount)
	as := services.NewAccrualService(cfg.AccrualAddr)

	app := &App{
		cfg:            cfg,
		router:         router,
		storage:        storage,
		accrualService: as,
		pool:           wp,
	}

	return app, nil
}

func (app *App) Run() {
	userHandler := handlers.NewUserHandler(app.storage, app.accrualService, app.pool)

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

	app.pool.Stop()
}
