package repository

import (
    "context"

    "TransactionSystem/internal/models"

    "github.com/jackc/pgx/v4/pgxpool"
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


func (tr *TransactionRepository) GetTransactionById(ctx context.Context, id int64) (*Transaction, error) {
    query := `SELECT id, from_wallet, to_wallet, amount, created_at 
    		  FROM "TransactionSystem".transactions WHERE id = $1`

    var transaction Transaction

    err := tr.db.QueryRow(ctx, query, id).Scan(
        &transaction.Id,
        &transaction.From,
        &transaction.To,
        &transaction.Amount,
        &transaction.CreatedAt,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("transaction with id %v not found: %w", id, err)
        }
        return nil, fmt.Errorf("failed to find transaction with id %v: %w", id, err)
    }

    return &transaction, nil
}

func (tr *TransactionRepository) GetTransactionByInfo(ctx context.Context, from, to string, createdAt time.Time) (*Transaction, error) {
	query := `SELECT id, from_wallet, to_wallet, amount, created_at
			  FROM "TransactionSystem".transactions 
			  WHERE from_wallet = $1 AND to_wallet = $2 AND created_at = $3`

	var transaction Transaction

	createdAt = createdAt.Truncate(time.Second) // чтобы не было проблем с миллисекундами

	err := tr.db.QueryRow(ctx, query, from, to, createdAt).Scan(
		&transaction.Id,
		&transaction.From,
		&transaction.To,
		&transaction.Amount,
		&transaction.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
		    return nil, fmt.Errorf("transaction not found for from_wallet %v, to_wallet %v at %v: %w", from, to, createdAt, err)
		}
		return nil, fmt.Errorf("transaction not found for from_wallet %v, to_wallet %v at %v: %w", from, to, createdAt, err)
	}

	return &transaction, nil
}
	
func (tr *TransactionRepository) RemoveTransaction(ctx context.Context, id int64) error {
	query := `DELETE FROM "TransactionSystem".transactions WHERE id = $1`

	_, err := tr.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction with id %v: %w", id, err)
	}

	return nil
}

func (tr *TransactionRepository) GetLastTransactions(ctx context.Context, limit int) ([]Transaction, error) {
    query := `SELECT id, from_wallet, to_wallet, amount, created_at 
              FROM "TransactionSystem".transactions 
              ORDER BY created_at DESC
              LIMIT $1`

    rows, err := tr.db.Query(ctx, query, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to get last transactions: %w", err)
    }
    defer rows.Close()

    transactions := make([]Transaction, 0, limit)

    for rows.Next() {
        var t Transaction

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
