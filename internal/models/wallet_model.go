package models 

import (
	"time"
)

type Wallet struct {
	Address   string    `json:"address"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}