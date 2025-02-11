package repository

import (
    "context"
    "fmt"
    "time"

    "TransactionSystem/internal/models"

    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/jackc/pgx/v4"
)

// Менеджер для транзакций
type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (tr *TransactionRepository) CreateTransaction(ctx context.Context, from, to string, amount float64) (int64, error) {
	query := `INSERT INTO "TransactionSystem".transactions (from_wallet, to_wallet, amount) 
              VALUES ($1, $2, $3) RETURNING id`

	var transactionId int64

	err := tr.db.QueryRow(ctx, query, from, to, amount).Scan(&transactionId)
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transactionId, nil
}

func (tr *TransactionRepository) ExecuteTransfer(ctx context.Context, from, to string, balance_from, balance_to, amount float64) error {
	tx, err := tr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем балансы кошельков
    updateQuery := `
        UPDATE "TransactionSystem".wallets 
        SET balance = CASE 
            WHEN address = $1 THEN CAST($3 AS numeric)
            WHEN address = $2 THEN CAST($4 AS numeric) 
        END 
        WHERE address IN ($1, $2)`

	_, err = tx.Exec(ctx, updateQuery, from, to, balance_from, balance_to)
    if err != nil {
        return fmt.Errorf("balance update failed: %w", err)
    }

	// Создаем запись о транзакции
	_, err = tx.Exec(ctx,
		`INSERT INTO "TransactionSystem".transactions 
		(from_wallet, to_wallet, amount) 
		VALUES ($1, $2, $3)`,
		from, to, amount,
	)
	if err != nil {
		return fmt.Errorf("transaction record failed: %w", err)
	}

	return tx.Commit(ctx)
}


func (tr *TransactionRepository) GetTransactionById(ctx context.Context, id int64) (*models.Transaction, error) {
    query := `SELECT id, from_wallet, to_wallet, amount, created_at 
    		  FROM "TransactionSystem".transactions WHERE id = $1`

    var t models.Transaction

    err := tr.db.QueryRow(ctx, query, id).Scan(
        &t.Id,
        &t.From,
        &t.To,
        &t.Amount,
        &t.CreatedAt,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("transaction with id %v not found: %w", id, err)
        }
        return nil, fmt.Errorf("failed to find transaction with id %v: %w", id, err)
    }

    return &t, nil
}
	
func (tr *TransactionRepository) GetTransactionByInfo(ctx context.Context, from, to string, createdAt time.Time) (*models.Transaction, error) {
    query := `SELECT id, from_wallet, to_wallet, amount, created_at
              FROM "TransactionSystem".transactions 
              WHERE from_wallet = $1 AND to_wallet = $2 AND created_at = $3`

    var t models.Transaction

    err := tr.db.QueryRow(ctx, query, from, to, createdAt).Scan(
        &t.Id,
        &t.From,
        &t.To,
        &t.Amount,
        &t.CreatedAt,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("transaction not found for from_wallet %v, to_wallet %v at %v: %w", from, to, createdAt, err)
        }
        return nil, fmt.Errorf("transaction not found for from_wallet %v, to_wallet %v at %v: %w", from, to, createdAt, err)
    }

    return &t, nil
}

func (tr *TransactionRepository) RemoveTransaction(ctx context.Context, id int64) error {
	query := `DELETE FROM "TransactionSystem".transactions WHERE id = $1`

	_, err := tr.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction with id %v: %w", id, err)
	}

	return nil
}

func (tr *TransactionRepository) GetLastTransactions(ctx context.Context, limit int) ([]models.Transaction, error) {
    query := `SELECT id, from_wallet, to_wallet, amount, created_at 
              FROM "TransactionSystem".transactions 
              ORDER BY created_at DESC
              LIMIT $1`

    rows, err := tr.db.Query(ctx, query, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to get last transactions: %w", err)
    }
    defer rows.Close()

    transactions := make([]models.Transaction, 0, limit)

    for rows.Next() {
        var t models.Transaction

        if err := rows.Scan(&t.Id, &t.From, &t.To, &t.Amount, &t.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan transaction: %w", err)
        }

        transactions = append(transactions, t)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error while fetching rows: %w", err)
    }

    return transactions, nil
}
