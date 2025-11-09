package main

import (
	"time"
)

type LoginRequest struct {
	Number   int64  `json:"account_number"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccount int     `json:"to_account"`
	Amount    float64 `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  string    `json:"-"`
	Number    int64     `json:"account_number"`
	Balance   float64   `json:"account_balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName string, password *string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Password:  *password,
		CreatedAt: time.Now().UTC(),
	}
}
