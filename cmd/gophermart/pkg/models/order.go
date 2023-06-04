package models

import (
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/helpers"
)

type Order struct {
	ID         uint                `json:"id"`
	UserID     uint                `json:"user_id"`
	Number     string              `json:"number"`
	Status     string              `json:"status"`
	Accrual    *float64            `json:"accrual,omitempty"`
	UploadedAt helpers.RFC3339Time `json:"uploaded_at"`
}
