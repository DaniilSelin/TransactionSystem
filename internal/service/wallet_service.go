package service

import (
	"context"
	"fmt"

	"TransactionSystem/internal/repository"
	"TransactionSystem/internal/models"
	
	"github.com/google/uuid"
)

type WalletService struct {
	walletRepo *repository.WalletRepository
}

func NewWalletService(walletRepo *repository.WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

func (ws *WalletService) CreateWallet(ctx context.Context) (string, error) {
	address := uuid.New().String()
	if err := ws.walletRepo.CreateWallet(ctx, address, 0.0); err != nil {
		return "", fmt.Errorf("failed to create wallet: %w", err)
	}
	return address, nil
}

func (ws *WalletService) GetBalance(ctx context.Context, address string) (float64, error) {
	balance, err := ws.walletRepo.GetWalletBalance(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance for wallet %s: %w", address, err)
	}
	return balance, nil
}

func (ws *WalletService) GetWallet(ctx context.Context, address string) (*models.Wallet, error) {
	wallet, err := ws.walletRepo.GetWallet(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet %s: %w", address, err)
	}
	return wallet, nil
}

func (ws *WalletService) UpdateBalance(ctx context.Context, address string, newBalance float64) error {
	if newBalance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}
	return ws.walletRepo.UpdateWalletBalabnce(ctx, address, newBalance)
}

func (ws *WalletService) RemoveWallet(ctx context.Context, address string) error {
	return ws.walletRepo.RemoveWallet(ctx, address)
}