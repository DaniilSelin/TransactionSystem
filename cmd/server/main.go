package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"TransactionSystem/api"
	"TransactionSystem/config"
	"TransactionSystem/internal/database"
	"TransactionSystem/internal/repository"
	"TransactionSystem/internal/service"
)

func main() {
	// 1. Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Подключаемся к БД
	dbPool, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbPool.Close()

	// 3. Запускаем миграции
	ctx := context.Background()
	err = database.RunMigrations(ctx, dbPool)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 4. Инициализируем репозитории
	transactionRepo := repository.NewTransactionRepository(dbPool)
	walletRepo := repository.NewWalletRepository(dbPool)

	// 5. Инициализируем сервисы
	transactionService := service.NewTransactionService(transactionRepo, walletRepo)
	walletService := service.NewWalletService(walletRepo)

	// 5.5. Создаем, при необходимости, начальные 10 кошельков
	if flagEmpty, err := walletService.IsEmpty(ctx); err != nil {
		log.Fatalf("Failed to check if wallets table is empty: %v", err)
	} else if flagEmpty {
		for i := 0; i < 10; i++ {
			address, err := walletService.CreateWallet(ctx, 100)
			if err != nil {
				log.Fatalf("Failed to create initial wallets: %v", err)
			}
			log.Printf("Created wallet %s\n", address)
		}
	}

	// 6. Создаём роутер
	router := api.NewRouter(transactionService, walletService)

	// 7. Запускаем сервер
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s...", addr)

	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
