CREATE SCHEMA IF NOT EXISTS "TransactionSystem";

CREATE TABLE IF NOT EXISTS "TransactionSystem".wallets (
    id SERIAL PRIMARY KEY,
    address TEXT UNIQUE NOT NULL,
    balance DECIMAL(18, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
