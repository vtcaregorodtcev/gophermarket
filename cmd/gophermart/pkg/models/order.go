package models

import (
	"time"
)

type Order struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
