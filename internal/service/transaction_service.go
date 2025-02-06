package service

import ()

func SendMoney(from, to string, amount float64) error

func GetLastTransactions(limit int) ([]Transaction, error)