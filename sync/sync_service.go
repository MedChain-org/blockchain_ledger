package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ankit/blockchain_ledger/blockchain"
	"github.com/ankit/blockchain_ledger/storage"
)

// SyncService handles automatic synchronization between local storage and Supabase
type SyncService struct {
	Storage       *storage.DataStorage
	Blockchain    *blockchain.BlockchainService
	SyncInterval  time.Duration
	StopChan      chan struct{}
	WaitGroup     sync.WaitGroup
	IsRunning     bool
	SyncLock      sync.Mutex
	LastSyncTime  time.Time
	SyncLogDir    string
	SyncStatusMap map[string]time.Time // Maps table names to last sync time
}

// SyncStatus represents the status of a synchronization operation
type SyncStatus struct {
	Table           string    `json:"table"`
	LastSync        time.Time `json:"last_sync"`
	RecordsSent     int       `json:"records_sent"`
	RecordsReceived int       `json:"records_received"`
	Status          string    `json:"status"`
	Error           string    `json:"error,omitempty"`
}

// NewSyncService creates a new synchronization service
func NewSyncService(dataStorage *storage.DataStorage, blockchain *blockchain.BlockchainService, syncInterval time.Duration) (*SyncService, error) {
	// Create sync logs directory if it doesn't exist
	syncLogDir := "sync_logs"
	if err := os.MkdirAll(syncLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sync logs directory: %v", err)
	}

	return &SyncService{
		Storage:       dataStorage,
		Blockchain:    blockchain,
		SyncInterval:  syncInterval,
		StopChan:      make(chan struct{}),
		IsRunning:     false,
		SyncLogDir:    syncLogDir,
		SyncStatusMap: make(map[string]time.Time),
	}, nil
}

// Start begins the synchronization service
func (s *SyncService) Start() error {
	s.SyncLock.Lock()
	defer s.SyncLock.Unlock()

	if s.IsRunning {
		return fmt.Errorf("sync service is already running")
	}

	s.IsRunning = true
	s.WaitGroup.Add(1)

	go func() {
		defer s.WaitGroup.Done()
		ticker := time.NewTicker(s.SyncInterval)
		defer ticker.Stop()

		// Perform initial sync
		s.performSync()

		for {
			select {
			case <-ticker.C:
				s.performSync()
			case <-s.StopChan:
				log.Println("Stopping sync service")
				return
			}
		}
	}()

	log.Printf("Sync service started with interval: %v", s.SyncInterval)
	return nil
}

// Stop stops the synchronization service
func (s *SyncService) Stop() error {
	s.SyncLock.Lock()
	defer s.SyncLock.Unlock()

	if !s.IsRunning {
		return fmt.Errorf("sync service is not running")
	}

	close(s.StopChan)
	s.WaitGroup.Wait()
	s.IsRunning = false

	log.Println("Sync service stopped")
	return nil
}

// performSync performs the actual synchronization
func (s *SyncService) performSync() {
	// Only log when manually triggered or during initial sync
	if s.LastSyncTime.IsZero() {
		log.Println("Starting initial synchronization process")
	}

	// Tables to synchronize
	tables := []string{"drugs", "shipments"}

	for _, table := range tables {
		status := s.syncTable(table)
		s.logSyncStatus(status)

		// Update last sync time for this table
		if status.Status == "success" {
			s.SyncStatusMap[table] = time.Now()
		}
	}

	s.LastSyncTime = time.Now()

	// Only log when manually triggered or during initial sync
	if s.LastSyncTime.IsZero() {
		log.Println("Synchronization process completed")
	}
}

// syncTable synchronizes a specific table
func (s *SyncService) syncTable(tableName string) SyncStatus {
	// Only log when there are changes
	logChanges := false

	status := SyncStatus{
		Table:    tableName,
		LastSync: time.Now(),
		Status:   "success",
	}

	// Get last sync time for this table
	lastSyncTime, exists := s.SyncStatusMap[tableName]
	if !exists {
		lastSyncTime = time.Time{} // If never synced, use zero time
	}

	// 1. Pull changes from Supabase
	recordsReceived, err := s.pullChangesFromSupabase(tableName)
	if err != nil {
		status.Status = "error"
		status.Error = fmt.Sprintf("Failed to pull changes from Supabase: %v", err)
		log.Printf("Error syncing %s: %v", tableName, err)
		return status
	}
	status.RecordsReceived = recordsReceived

	if recordsReceived > 0 {
		logChanges = true
	}

	// 2. Push local changes to Supabase
	recordsSent, err := s.pushChangesToSupabase(tableName, lastSyncTime)
	if err != nil {
		status.Status = "error"
		status.Error = fmt.Sprintf("Failed to push changes to Supabase: %v", err)
		log.Printf("Error pushing changes for %s: %v", tableName, err)
		return status
	}
	status.RecordsSent = recordsSent

	if recordsSent > 0 {
		logChanges = true
	}

	// Only log when there are actual changes
	if logChanges {
		log.Printf("Sync completed for %s: Received %d records, Sent %d records",
			tableName, recordsReceived, recordsSent)
	}

	return status
}

// pullChangesFromSupabase pulls changes from Supabase and creates blockchain transactions
func (s *SyncService) pullChangesFromSupabase(table string) (int, error) {
	// Get records from Supabase
	records, err := s.Storage.Supabase.Select(table, "*", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch records from Supabase: %v", err)
	}

	// Only log if there are records to process
	if len(records) > 0 {
		log.Printf("Received %d records for table %s", len(records), table)
	}

	// Count of records that need processing
	processedCount := 0

	// Process each record
	for _, record := range records {
		// Skip records that already have a valid blockchain transaction
		if txID, ok := record["blockchain_tx_id"].(string); ok && txID != "pending" {
			continue
		}

		// Increment processed count
		processedCount++

		// Create blockchain transaction for drug records
		if table == "drugs" {
			// Create a copy of the record to avoid modifying the original
			drugData := make(map[string]interface{})
			for k, v := range record {
				drugData[k] = v
			}

			// Create blockchain transaction
			txHash, err := s.Blockchain.CreateTransaction("drug", drugData)
			if err != nil {
				log.Printf("Failed to create blockchain transaction for drug %s: %v", drugData["drug_id"], err)
				continue
			}

			// Update Supabase record with blockchain hash
			updateData := map[string]interface{}{
				"blockchain_tx_id": txHash,
			}
			_, err = s.Storage.Supabase.Update("drugs", drugData["drug_id"].(string), updateData)
			if err != nil {
				log.Printf("Failed to update Supabase record with blockchain hash: %v", err)
			} else {
				log.Printf("Created blockchain transaction for drug %s: %s", drugData["drug_id"], txHash)
			}
		}
	}

	return processedCount, nil
}

// createShipmentTransaction creates a blockchain transaction for a shipment record
func (s *SyncService) createShipmentTransaction(shipmentData map[string]interface{}) string {
	// Skip if already has a valid blockchain transaction
	if txID, exists := shipmentData["blockchain_tx_id"].(string); exists && txID != "" && txID != "pending" {
		return txID
	}

	// Create a copy of the shipment data to avoid modifying the original
	txData := make(map[string]interface{})
	for k, v := range shipmentData {
		txData[k] = v
	}

	// Create blockchain transaction
	_, err := blockchain.NewShipmentTransaction(txData)
	if err != nil {
		log.Printf("Failed to create shipment transaction: %v", err)
		return ""
	}

	// Generate transaction hash
	txHash := blockchain.GenerateTransactionHash(txData)

	// Add transaction to blockchain
	if err := s.Storage.AddTransactionToBlockchain(txData, txHash); err != nil {
		log.Printf("Failed to add shipment transaction to blockchain: %v", err)
		return ""
	}

	log.Printf("Created blockchain transaction for shipment %s: %s",
		txData["shipment_id"], txHash)

	// Update the original shipment data with the transaction hash
	shipmentData["blockchain_tx_id"] = txHash

	return txHash
}

// pushChangesToSupabase pushes local changes to Supabase
func (s *SyncService) pushChangesToSupabase(tableName string, lastSyncTime time.Time) (int, error) {
	// In a real implementation, you'd track local changes and push them to Supabase
	// For this example, we'll assume there are no local changes to push

	// This would involve:
	// 1. Identifying local records that have been created or modified since the last sync
	// 2. Pushing those changes to Supabase
	// 3. Handling any conflicts that arise

	// For now, we'll just return 0 records sent
	return 0, nil
}

// logSyncStatus logs the synchronization status
func (s *SyncService) logSyncStatus(status SyncStatus) {
	// Create log file name with timestamp
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("%s-sync-%s.json", status.Table, timestamp)
	logFilePath := filepath.Join(s.SyncLogDir, logFileName)

	// Marshal status to JSON
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal sync status: %v", err)
		return
	}

	// Write to log file
	if err := os.WriteFile(logFilePath, data, 0644); err != nil {
		log.Printf("Failed to write sync log: %v", err)
	}
}

// GetSyncStatus returns the current synchronization status
func (s *SyncService) GetSyncStatus() map[string]interface{} {
	s.SyncLock.Lock()
	defer s.SyncLock.Unlock()

	statusMap := make(map[string]interface{})
	statusMap["is_running"] = s.IsRunning
	statusMap["last_sync"] = s.LastSyncTime
	statusMap["sync_interval"] = s.SyncInterval.String()

	tableStatus := make(map[string]interface{})
	for table, lastSync := range s.SyncStatusMap {
		tableStatus[table] = map[string]interface{}{
			"last_sync": lastSync,
		}
	}
	statusMap["tables"] = tableStatus

	return statusMap
}

// ForceSync forces an immediate synchronization
func (s *SyncService) ForceSync() error {
	if !s.IsRunning {
		return fmt.Errorf("sync service is not running")
	}

	go s.performSync()
	return nil
}

// createDrugTransaction creates a blockchain transaction for a drug record
func (s *SyncService) createDrugTransaction(drugData map[string]interface{}) string {
	// Skip if already has a valid blockchain transaction
	if txID, exists := drugData["blockchain_tx_id"].(string); exists && txID != "" && txID != "pending" {
		return txID
	}

	// Create a copy of the drug data to avoid modifying the original
	txData := make(map[string]interface{})
	for k, v := range drugData {
		txData[k] = v
	}

	// Create blockchain transaction
	_, err := blockchain.NewDrugTransaction(txData)
	if err != nil {
		log.Printf("Failed to create drug transaction: %v", err)
		return ""
	}

	// Generate transaction hash
	txHash := blockchain.GenerateTransactionHash(txData)

	// Add transaction to blockchain
	if err := s.Storage.AddTransactionToBlockchain(txData, txHash); err != nil {
		log.Printf("Failed to add drug transaction to blockchain: %v", err)
		return ""
	}

	log.Printf("Created blockchain transaction for drug %s: %s",
		txData["drug_id"], txHash)

	// Update the original drug data with the transaction hash
	drugData["blockchain_tx_id"] = txHash

	return txHash
}
