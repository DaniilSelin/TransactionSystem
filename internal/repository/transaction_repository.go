package repository

import ()

func CreateTransaction(from, to string, amount float64) error

func GetLastTransactions(limit int) ([]Transaction, error) 