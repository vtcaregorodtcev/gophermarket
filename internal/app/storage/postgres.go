package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/helpers"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

type Storager interface {
	Close() error
	GetUserByID(tx *sql.Tx, id uint) (*models.User, error)
	GetUserByLogin(tx *sql.Tx, login string) (*models.User, error)
	CreateUser(ctx context.Context, login, password string) (*models.User, error)
	GetOrderByNumber(tx *sql.Tx, orderNumber string) (*models.Order, error)
	CreateOrder(ctx context.Context, orderNumber string, userID uint) (*models.Order, error)
	UpdateOrderAccrualAndUserBalance(ctx context.Context, orderID uint, userID uint, accrualResp *services.CalcOrderAccrualResponse) error
	GetOrdersByUserID(userID uint) (*[](*models.Order), error)
	WithdrawBalance(ctx context.Context, userID uint, orderNumber string, withdrawalAmount float64) error
	GetUserWithdrawals(userID uint) (*[](*models.Withdrawal), error)
}

type Storage struct {
	db *sql.DB
}

func New(dbURI string) *Storage {
	baseDir, _ := os.Getwd()

	logger.Infof("using postress db at: %s", dbURI)

	file := filepath.Join(baseDir, "..", "..", "migrations", "init.sql")

	isDevMode := *helpers.GetStringEnv("DEV_MODE", helpers.StringPtr("false")) == "true"

	if !isDevMode {
		file = filepath.Join(baseDir, "migrations", "init.sql")
	}

	init, err := os.ReadFile(file)
	if err != nil {
		logger.Infof("init script is not found: %v", err)
	}

	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(string(init))
	if err != nil {
		panic(err)
	}

	return &Storage{db: db}
}

func (s *Storage) Close() error {
	return s.db.Close()
}
