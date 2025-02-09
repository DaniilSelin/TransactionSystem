package service_test

import (
	"context"
	"testing"
	"time"

	"TransactionSystem/internal/database"
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

type WalletServiceTestSuite struct {
	suite.Suite
	container  *postgres.PostgresContainer
	dbPool     *pgxpool.Pool
	service    *service.WalletService
	ctx        context.Context
}

func (suite *WalletServiceTestSuite) SetupSuite() {
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

	repo := repository.NewWalletRepository(pool)
	suite.service = service.NewWalletService(repo)
}

func (suite *WalletServiceTestSuite) TearDownSuite() {
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
	if suite.dbPool != nil {
		suite.dbPool.Close()
	}
}

func (suite *WalletServiceTestSuite) BeforeTest(_, _ string) {
	_, err := suite.dbPool.Exec(suite.ctx, `DELETE FROM "TransactionSystem".wallets`)
	assert.NoError(suite.T(), err)
}

func TestWalletService(t *testing.T) {
	suite.Run(t, new(WalletServiceTestSuite))
}

func (suite *WalletServiceTestSuite) createTestWallet(balance float64) string {
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

func (suite *WalletServiceTestSuite) TestCreateWallet() {
	t := suite.T()
	ctx := context.Background()

	address, err := suite.service.CreateWallet(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, address)

	// Проверяем существование в БД
	var count int
	err = suite.dbPool.QueryRow(ctx, 
		`SELECT COUNT(*) FROM "TransactionSystem".wallets WHERE address = $1`, address).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func (suite *WalletServiceTestSuite) TestGetBalance() {
	t := suite.T()
	ctx := context.Background()

	address := suite.createTestWallet(100.0)

	balance, err := suite.service.GetBalance(ctx, address)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, balance)

	// Несуществующий кошелек
	_, err = suite.service.GetBalance(ctx, "invalid_address")
	assert.ErrorContains(t, err, "not found")
}

func (suite *WalletServiceTestSuite) TestGetWallet() {
	t := suite.T()
	ctx := context.Background()

	address := suite.createTestWallet(200.0)

	wallet, err := suite.service.GetWallet(ctx, address)
	assert.NoError(t, err)

	assert.Equal(t, address, wallet.Address)
	assert.Equal(t, 200.0, wallet.Balance)
	assert.WithinDuration(t, time.Now(), wallet.CreatedAt, 2*time.Second)
}

func (suite *WalletServiceTestSuite) TestUpdateBalance() {
	t := suite.T()
	ctx := context.Background()

	address := suite.createTestWallet(50.0)

	tests := []struct {
		name        string
		newBalance  float64
		expectError bool
	}{
		{"Positive value", 100.0, false},
		{"Negative value", -10.0, true},
		{"Zero value", 0.0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := suite.service.UpdateBalance(ctx, address, tc.newBalance)
			if tc.expectError {
				assert.ErrorContains(t, err, "cannot be negative")
			} else {
				assert.NoError(t, err)
				balance, _ := suite.service.GetBalance(ctx, address)
				assert.Equal(t, tc.newBalance, balance)
			}
		})
	}
}

func (suite *WalletServiceTestSuite) TestRemoveWallet() {
	t := suite.T()
	ctx := context.Background()

	address := suite.createTestWallet(0.0)

	err := suite.service.RemoveWallet(ctx, address)
	assert.NoError(t, err)

	// Проверяем удаление
	var count int
	err = suite.dbPool.QueryRow(ctx, 
		`SELECT COUNT(*) FROM "TransactionSystem".wallets WHERE address = $1`, address).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func (suite *WalletServiceTestSuite) TestRemoveWallet_NonExistent() {
	t := suite.T()
	ctx := context.Background()

	err := suite.service.RemoveWallet(ctx, "non_existent_address")
	assert.ErrorContains(t, err, "wallet with address non_existent_address not found")
}
