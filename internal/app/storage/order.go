package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vtcaregorodtcev/gophermarket/internal/app/services"
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

func (s *Storage) GetOrderByNumber(orderNumber string) (*models.Order, error) {
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

	row := s.db.QueryRow(query, orderNumber)

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

func (s *Storage) UpdateOrderAccrualAndUserBalance(ctx context.Context, orderID uint, userID uint, accrualResp *services.CalcOrderAccrualResponse) error {
	s.log.Info().Msgf("UpdateOrderAccrualAndUserBalance params: orderID: %d, userID: %d", orderID, userID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

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

	s.log.Info().Msg("UpdateOrderAccrualAndUserBalance: Successfully updated")

	return nil
}

func (s *Storage) WithdrawBalance(ctx context.Context, userID uint, orderNumber string, withdrawalAmount float64) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	s.log.Info().Msgf(
		"WithdrawBalance with: orderNumber: %s, withdrawalAmount: %f, userID: %d, userBalance: %f",
		orderNumber, withdrawalAmount, userID, user.Balance)

	if user.Balance < withdrawalAmount {
		return ErrInsufficientBalance
	}

	existingOrder, err := s.GetOrderByNumber(orderNumber)
	if err != nil {
		return err
	}
	if existingOrder != nil {
		return ErrOrderAlreadyExists
	}

	newBalance := user.Balance - withdrawalAmount

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `INSERT INTO orders (number, user_id, status) VALUES ($1, $2, $3) RETURNING id`
	var orderID uint
	err = tx.QueryRowContext(ctx, query, orderNumber, userID, models.PROCESSED).Scan(&orderID)
	if err != nil {
		return err
	}

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

	query = `INSERT INTO withdrawals (user_id, order_id, sum) VALUES ($1, $2, $3) RETURNING id`
	var wID uint
	err = tx.QueryRowContext(ctx, query, userID, orderID, withdrawalAmount).Scan(&wID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	s.log.Info().Msg("WithdrawBalance: Successfully updated")

	return nil
}
