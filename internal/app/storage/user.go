package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

func (s *Storage) CreateUser(ctx context.Context, login, password string) (*models.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
	var id uint
	err = tx.QueryRowContext(ctx, query, login, password).Scan(&id)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       id,
		Login:    login,
		Password: password,
	}, nil
}

func (s *Storage) getUserBy(by, what string) (*models.User, error) {
	query := "SELECT id, login, password, balance, withdrawn FROM users WHERE " + by + " = $1"

	row := s.db.QueryRow(query, what)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Login, &user.Password, &user.Balance, &user.Withdrawn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err // Other error occurred
	}

	return user, nil
}

func (s *Storage) GetUserByLogin(login string) (*models.User, error) {
	return s.getUserBy("login", login)
}

func (s *Storage) GetUserByID(id uint) (*models.User, error) {
	return s.getUserBy("id", fmt.Sprint(id))
}

func (s *Storage) GetUserWithdrawals(userID uint) (*[](*models.Withdrawal), error) {
	withdrawals := make([]*models.Withdrawal, 0)

	rows, err := s.db.Query(`
		SELECT
			o.number,
			w.sum,
			w.processed_at
		FROM
			withdrawals w
		JOIN
			orders o ON w.order_id = o.id
		WHERE
			w.user_id = $1
		ORDER BY
			w.processed_at ASC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal models.Withdrawal
		if err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &withdrawals, nil
}
