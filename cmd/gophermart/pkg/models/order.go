package models

type Order struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	Number string `json:"number"`
}
