package storage

import (
	"context"
	"database/sql"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"
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

func (s *Storage) GetUserByLogin(login string) (*models.User, error) {
	query := "SELECT id, login, password, balance, withdrawn FROM users WHERE login = $1"

	row := s.db.QueryRow(query, login)

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

func (s *Storage) GetUserByID(id uint) (*models.User, error) {
	query := "SELECT id, login, password, balance, withdrawn FROM users WHERE id = $1"

	row := s.db.QueryRow(query, id)

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
