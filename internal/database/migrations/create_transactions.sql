CREATE TABLE IF NOT EXISTS "TransactionSystem".transactions (
    id SERIAL PRIMARY KEY,
    from_wallet TEXT NOT NULL,
    to_wallet TEXT NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_from FOREIGN KEY (from_wallet) REFERENCES "TransactionSystem".wallets(address),
    CONSTRAINT fk_to FOREIGN KEY (to_wallet) REFERENCES "TransactionSystem".wallets(address)
);
