CREATE SCHEMA IF NOT EXISTS "TransactionSystem";

CREATE TABLE IF NOT EXISTS "TransactionSystem".wallets (
    address TEXT PRIMARY KEY,
    balance DECIMAL(18, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
