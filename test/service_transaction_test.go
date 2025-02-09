package service_test

import (
	"context"
	"testing"
	"time"

	"TransactionSystem/internal/database"
	"TransactionSystem/internal/models"
	"TransactionSystem/internal/repository"
	"TransactionSystem/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TransactionServiceTestSuite struct {
	suite.Suite
	container  *postgres.PostgresContainer
	dbPool     *pgxpool.Pool
	service    *service.TransactionService
	ctx        context.Context
}

func (suite *TransactionServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	container, err := postgres.RunContainer(
		suite.ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.container = container

	connStr, err := container.ConnectionString(suite.ctx, "sslmode=disable")
	if err != nil {
		suite.T().Fatal(err)
	}

	pool, err := pgxpool.Connect(suite.ctx, connStr)
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.dbPool = pool

	err = database.RunMigrations(suite.ctx, pool)
	if err != nil {
		suite.T().Fatal(err)
	}

	walletRepo := repository.NewWalletRepository(pool)
	transactionRepo := repository.NewTransactionRepository(pool)
	suite.service = service.NewTransactionService(transactionRepo, walletRepo)
}

func (suite *TransactionServiceTestSuite) TearDownSuite() {
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
	if suite.dbPool != nil {
		suite.dbPool.Close()
	}
}

func (suite *TransactionServiceTestSuite) BeforeTest(_, _ string) {
	_, err := suite.dbPool.Exec(suite.ctx, `
		TRUNCATE TABLE "TransactionSystem".wallets CASCADE;
		TRUNCATE TABLE "TransactionSystem".transactions CASCADE;
	`)
	assert.NoError(suite.T(), err)
}

func TestTransactionService(t *testing.T) {
	suite.Run(t, new(TransactionServiceTestSuite))
}

func (suite *TransactionServiceTestSuite) createTestWallet(balance float64) string {
	address := uuid.New().String()
	_, err := suite.dbPool.Exec(suite.ctx, `
		INSERT INTO "TransactionSystem".wallets (address, balance)
		VALUES ($1, $2)
	`, address, balance)
	if err != nil {
		suite.T().Fatal(err)
	}
	return address
}

func (suite *TransactionServiceTestSuite) TestSendMoney_Success() {
	ctx := context.Background()
	
	from := suite.createTestWallet(100.0)
	to := suite.createTestWallet(50.0)

	err := suite.service.SendMoney(ctx, from, to, 30.0)
	assert.NoError(suite.T(), err)

	var fromBalance, toBalance float64
	err = suite.dbPool.QueryRow(ctx, 
		`SELECT balance FROM "TransactionSystem".wallets WHERE address = $1`, from).Scan(&fromBalance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 70.0, fromBalance)

	err = suite.dbPool.QueryRow(ctx, 
		`SELECT balance FROM "TransactionSystem".wallets WHERE address = $1`, to).Scan(&toBalance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 80.0, toBalance)

	var transaction models.Transaction
	err = suite.dbPool.QueryRow(ctx, `
		SELECT from_wallet, to_wallet, amount 
		FROM "TransactionSystem".transactions
		WHERE from_wallet = $1 AND to_wallet = $2
	`, from, to).Scan(
		&transaction.From,
		&transaction.To,
		&transaction.Amount,
	)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 30.0, transaction.Amount)
}

func (suite *TransactionServiceTestSuite) TestSendMoney_InvalidParameters() {
	ctx := context.Background()
	from := suite.createTestWallet(100.0)

	tests := []struct {
		name    string
		from    string
		to      string
		amount  float64
		wantErr string
	}{
		{"Same sender and receiver", from, from, 10.0, "cannot be the same"},
		{"Negative amount", from, "any", -5.0, "greater than zero"},
		{"Zero amount", from, "any", 0.0, "greater than zero"},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.service.SendMoney(ctx, tc.from, tc.to, tc.amount)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func (suite *TransactionServiceTestSuite) TestSendMoney_InsufficientFunds() {
	ctx := context.Background()
	from := suite.createTestWallet(50.0)
	to := suite.createTestWallet(0.0)

	err := suite.service.SendMoney(ctx, from, to, 100.0)
	assert.ErrorContains(suite.T(), err, "insufficient funds")
}

func (suite *TransactionServiceTestSuite) TestGetLastTransactions() {
	ctx := context.Background()
	from := suite.createTestWallet(200.0)
	to := suite.createTestWallet(0.0)

	for i := 0; i < 5; i++ {
		err := suite.service.SendMoney(ctx, from, to, 10.0)
		assert.NoError(suite.T(), err)
	}

	transactions, err := suite.service.GetLastTransactions(ctx, 3)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), transactions, 3)

	for _, t := range transactions {
		assert.Equal(suite.T(), from, t.From)
		assert.Equal(suite.T(), to, t.To)
		assert.Equal(suite.T(), 10.0, t.Amount)
	}
}

func (suite *TransactionServiceTestSuite) TestGetTransactionById() {
	ctx := context.Background()
	from := suite.createTestWallet(100.0)
	to := suite.createTestWallet(0.0)

	err := suite.service.SendMoney(ctx, from, to, 50.0)
	assert.NoError(suite.T(), err)

	var transactionID int64
	err = suite.dbPool.QueryRow(ctx, `
		SELECT id FROM "TransactionSystem".transactions 
		WHERE from_wallet = $1 AND to_wallet = $2
	`, from, to).Scan(&transactionID)
	assert.NoError(suite.T(), err)

	transaction, err := suite.service.GetTransactionById(ctx, transactionID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), transactionID, transaction.Id)
	assert.Equal(suite.T(), 50.0, transaction.Amount)
}

func (suite *TransactionServiceTestSuite) TestRemoveTransaction() {
	ctx := context.Background()
	from := suite.createTestWallet(100.0)
	to := suite.createTestWallet(0.0)

	err := suite.service.SendMoney(ctx, from, to, 30.0)
	assert.NoError(suite.T(), err)

	var transactionID int64
	err = suite.dbPool.QueryRow(ctx, `
		SELECT id FROM "TransactionSystem".transactions 
		WHERE from_wallet = $1 AND to_wallet = $2
	`, from, to).Scan(&transactionID)
	assert.NoError(suite.T(), err)

	err = suite.service.RemoveTransaction(ctx, transactionID)
	assert.NoError(suite.T(), err)

	_, err = suite.service.GetTransactionById(ctx, transactionID)
	assert.ErrorContains(suite.T(), err, "not found")
}