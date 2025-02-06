package repository

import ()

func CreateWallet(address string, balance float64) error

func GetWalletBalance(address string) (float64, error)