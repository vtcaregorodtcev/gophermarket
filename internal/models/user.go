package models

type User struct {
	ID        uint    `json:"id"`
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
