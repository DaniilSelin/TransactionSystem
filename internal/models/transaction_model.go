package models

import (
	"time"
)

type Transaction struct {
	ID        string    `json:"id"`
	From      string   `json:"from,omitempty"`
	To        string   `json:"to,omitempty"`  
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}