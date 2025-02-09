package service

import (
    "context"
    "errors"
    "fmt"
    "time"

    "TransactionSystem/internal/models"
    "TransactionSystem/internal/repository"
)

type TransactionService struct {
    transactionRepo *repository.TransactionRepository
    walletRepo      *repository.WalletRepository
}

func NewTransactionService(tr *repository.TransactionRepository, wr *repository.WalletRepository) *TransactionService {
    return &TransactionService{
        transactionRepo: tr,
        walletRepo:      wr,
    }
}

func (ts *TransactionService) SendMoney(ctx context.Context, from, to string, amount float64) error {
    if from == to {
        return errors.New("sender and receiver cannot be the same")
    }
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    fromBalance, err := ts.walletRepo.GetWalletBalance(ctx, from)
    if err != nil {
        return fmt.Errorf("failed to get sender balance: %w", err)
    }
    if fromBalance < amount {
        return errors.New("insufficient funds")
    }

    toBalance, err := ts.walletRepo.GetWalletBalance(ctx, to)
    if err != nil {
        return fmt.Errorf("failed to get receiver balance: %w", err)
    }

    newFromBalance := fromBalance - amount
    newToBalance := toBalance + amount

    // проиводим транзакцию
    return ts.transactionRepo.ExecuteTransfer(
        ctx,
        from,
        to,
        newFromBalance,
        newToBalance,
        amount,
    )
}

func (ts *TransactionService) GetLastTransactions(ctx context.Context, limit int) ([]models.Transaction, error) {
    if limit <= 0 {
        return nil, errors.New("limit must be greater than zero")
    }

    transactions, err := ts.transactionRepo.GetLastTransactions(ctx, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
    }

    return transactions, nil
}

func (ts *TransactionService) GetTransactionById(ctx context.Context, id int64) (*models.Transaction, error) {
    transaction, err := ts.transactionRepo.GetTransactionById(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction: %w", err)
    }
    return transaction, nil
}

func (ts *TransactionService) GetTransactionByInfo(ctx context.Context, from, to string, createdAt time.Time) (*models.Transaction, error) {
    transaction, err := ts.transactionRepo.GetTransactionByInfo(ctx, from, to, createdAt)
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction by info: %w", err)
    }
    return transaction, nil
}

func (ts *TransactionService) RemoveTransaction(ctx context.Context, id int64) error {
    if err := ts.transactionRepo.RemoveTransaction(ctx, id); err != nil {
        return fmt.Errorf("failed to remove transaction: %w", err)
    }
    return nil
}
