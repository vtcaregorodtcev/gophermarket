package models

import (
	"github.com/vtcaregorodtcev/gophermarket/internal/helpers"
)

type OrderStatus string

const (
	NEW        OrderStatus = "NEW"
	INVALID    OrderStatus = "INVALID"
	PROCESSING OrderStatus = "PROCESSING"
	PROCESSED  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         uint                `json:"id"`
	UserID     uint                `json:"user_id,omitempty"`
	Number     string              `json:"number"`
	Status     OrderStatus         `json:"status"`
	Accrual    *float64            `json:"accrual,omitempty"`
	UploadedAt helpers.RFC3339Time `json:"uploaded_at"`
}
