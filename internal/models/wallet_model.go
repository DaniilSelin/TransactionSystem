package models 

import ()

type Wallet struct {
	Address   string    `json:"address"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}