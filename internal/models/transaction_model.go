package models

import (
	"time"
)

type Transaction struct {
	Id       int64    `json:"id"`
	From      string   `json:"from"`
	To        string   `json:"to"`  
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}