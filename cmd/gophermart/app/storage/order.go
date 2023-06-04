package storage

import (
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"
)

func (s *Storage) GetOrdersByUserID(userID uint) ([]models.Order, error) {
	var orders []models.Order

	rows, err := s.db.Query("SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
