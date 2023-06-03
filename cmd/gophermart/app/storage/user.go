package storage

import (
	"database/sql"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"
)

func (s *Storage) CreateUser(user *models.User) error {
	// Implementation to create a user in the database
	return nil
}

func (s *Storage) GetUserByUsername(username string) (*models.User, error) {
	query := "SELECT id, username, password FROM users WHERE username = $1"

	row := s.db.QueryRow(query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err // Other error occurred
	}

	return user, nil
}
