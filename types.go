package main

import (
	"math/rand"
	"time"
)

type TransferRequest struct {
	ToAccount int     `json:"to_account"`
	Amount    float64 `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Number    int64     `json:"account_number"`
	Balance   int64     `json:"account_balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		ID:        rand.Intn(1000),
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(10000000)),
		CreatedAt: time.Now().UTC(),
	}
}
