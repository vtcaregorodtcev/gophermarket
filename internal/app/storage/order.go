package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrOrderAlreadyExists = errors.New("order already exists")

func (s *Storage) GetOrdersByUserID(userID uint) (*[](*models.Order), error) {
	orders := make([]*models.Order, 0)

	rows, err := s.db.Query(`
		SELECT
			id,
			user_id,
			number,
			status,
			accrual,
			uploaded_at
		FROM
			orders
		WHERE
			user_id = $1
		ORDER BY
			uploaded_at ASC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &orders, nil
}

func (s *Storage) GetOrderByNumber(tx *sql.Tx, orderNumber string) (*models.Order, error) {
	query := `
		SELECT
			id,
			user_id,
			number,
			status,
			accrual,
			uploaded_at
		FROM
			orders
		WHERE
			number = $1`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, orderNumber)
	} else {
		row = s.db.QueryRow(query, orderNumber)
	}

	var order models.Order
	if err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Order not found
		}
		return nil, err // Other error occurred
	}

	return &order, nil
}

func (s *Storage) CreateOrder(ctx context.Context, orderNumber string, userID uint) (*models.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	query := `INSERT INTO orders (number, user_id, status) VALUES ($1, $2, $3) RETURNING id`
	var id uint
	err = tx.QueryRowContext(ctx, query, orderNumber, userID, models.NEW).Scan(&id)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &models.Order{
		ID:     id,
		Number: orderNumber,
		Status: models.NEW,
	}, nil
}

func (s *Storage) UpdateOrderStatus(ctx context.Context, orderID uint, status models.OrderStatus) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	s.LockOrders(ctx, tx, orderID)

	oStmt, err := tx.PrepareContext(ctx, "UPDATE orders SET status = $1 WHERE id = $2")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer oStmt.Close()

	_, err = oStmt.ExecContext(ctx, status, orderID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (s *Storage) lock(ctx context.Context, tx *sql.Tx, ID uint, what string) error {
	smt, err := tx.PrepareContext(ctx, "SELECT id FROM "+what+" WHERE id = $1 FOR UPDATE")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer smt.Close()

	_, err = smt.ExecContext(ctx, ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (s *Storage) LockUsers(ctx context.Context, tx *sql.Tx, userID uint) error {
	return s.lock(ctx, tx, userID, "users")
}

func (s *Storage) LockOrders(ctx context.Context, tx *sql.Tx, orderID uint) error {
	return s.lock(ctx, tx, orderID, "orders")
}

func (s *Storage) UpdateOrderAccrualAndUserBalance(ctx context.Context, orderID uint, userID uint, accrualResp *services.CalcOrderAccrualResponse) error {
	logger.Infof("UpdateOrderAccrualAndUserBalance params: orderID: %d, userID: %d", orderID, userID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = s.LockOrders(ctx, tx, orderID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	oStmt, err := tx.PrepareContext(ctx, "UPDATE orders SET accrual = $1, status = $2 WHERE id = $3")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer oStmt.Close()

	_, err = oStmt.ExecContext(ctx, accrualResp.Accrual, accrualResp.Status, orderID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = s.LockUsers(ctx, tx, userID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	uStmt, err := tx.PrepareContext(ctx, "UPDATE users SET balance = balance + $1 WHERE id = $2")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer uStmt.Close()

	_, err = uStmt.ExecContext(ctx, accrualResp.Accrual, userID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	logger.Infof("UpdateOrderAccrualAndUserBalance: Successfully updated")

	return nil
}

func (s *Storage) WithdrawBalance(ctx context.Context, userID uint, orderNumber string, withdrawalAmount float64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	user, err := s.GetUserByID(tx, userID, true)
	if err != nil {
		return err
	}

	logger.Infof(
		"WithdrawBalance with: orderNumber: %s, withdrawalAmount: %f, userID: %d, userBalance: %f",
		orderNumber, withdrawalAmount, userID, user.Balance)

	if user.Balance < withdrawalAmount {
		return ErrInsufficientBalance
	}

	newBalance := user.Balance - withdrawalAmount

	uStmt, err := tx.PrepareContext(ctx, "UPDATE users SET balance = $1, withdrawn = withdrawn + $2 WHERE id = $3")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer uStmt.Close()

	_, err = uStmt.ExecContext(ctx, newBalance, withdrawalAmount, userID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	query := `INSERT INTO withdrawals (user_id, order_id, sum) VALUES ($1, $2, $3) RETURNING id`
	var wID uint
	err = tx.QueryRowContext(ctx, query, userID, orderNumber, withdrawalAmount).Scan(&wID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	logger.Infof("WithdrawBalance: Successfully updated")

	return nil
}
