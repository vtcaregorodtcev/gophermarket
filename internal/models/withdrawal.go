package models

import (
	"github.com/vtcaregorodtcev/gophermarket/internal/helpers"
)

type Withdrawal struct {
	OrderNumber string              `json:"order"`
	Sum         float64             `json:"sum"`
	ProcessedAt helpers.RFC3339Time `json:"processed_at"`
}
