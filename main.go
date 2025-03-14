package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ankit/blockchain_ledger/blockchain"
	"github.com/ankit/blockchain_ledger/handlers"
	"github.com/ankit/blockchain_ledger/manager"
	"github.com/ankit/blockchain_ledger/storage"
	"github.com/ankit/blockchain_ledger/sync"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	// Initialize storage
	dataStorage, err := storage.NewDataStorage()
	if err != nil {
		log.Fatalf("Failed to initialize data storage: %v", err)
	}

	ledgerStorage, err := storage.NewLedgerStorage()
	if err != nil {
		log.Fatalf("Failed to initialize ledger storage: %v", err)
	}

	// Initialize blockchain service
	blockchainService := blockchain.NewBlockchainService(dataStorage)

	// Initialize ledger manager
	ledgerManager := manager.NewLedgerManager(ledgerStorage, blockchainService)

	// Initialize sync service
	syncInterval := 1 * time.Minute // Default sync interval
	syncService, err := sync.NewSyncService(dataStorage, blockchainService, syncInterval)
	if err != nil {
		log.Fatalf("Failed to initialize sync service: %v", err)
	}

	// Initialize handlers
	handlers.SetupRoutes(ledgerManager, syncService)

	// Start sync service
	go syncService.Start()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start HTTP server
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
