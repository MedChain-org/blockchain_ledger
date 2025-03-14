package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ankit/blockchain_ledger/models"
	"github.com/ankit/blockchain_ledger/supabase"
)

// DataStorage handles all data storage operations
type DataStorage struct {
	Supabase       *supabase.Client
	DataDir        string
	WalDir         string
	BlockchainDir  string
	BlockchainFile string
}

// BlockchainLedger represents the structure of the blockchain ledger file
type BlockchainLedger struct {
	Blocks      []Block `json:"blocks"`
	LastUpdated string  `json:"last_updated"`
	BlockHeight int     `json:"block_height"`
}

// Block represents a single block in the blockchain
type Block struct {
	BlockHeight       int         `json:"block_height"`
	TxHash            string      `json:"tx_hash"`
	TxData            interface{} `json:"tx_data"`
	Timestamp         string      `json:"timestamp"`
	PreviousBlockHash string      `json:"previous_block_hash"`
}

// ConsistencyCheckResult represents the result of a consistency check
type ConsistencyCheckResult struct {
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Timestamp      string `json:"timestamp"`
	Recommendation string `json:"recommendation,omitempty"`
}

// NewDataStorage initializes a new DataStorage instance
func NewDataStorage() (*DataStorage, error) {
	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient(true) // Use service key
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %v", err)
	}

	// Create storage instance
	storage := &DataStorage{
		Supabase:       supabaseClient,
		DataDir:        "data_records",
		WalDir:         "wal_logs",
		BlockchainDir:  "blockchain_data",
		BlockchainFile: filepath.Join("blockchain_data", "blockchain_ledger.json"),
	}

	// Ensure directories exist
	dirs := []string{storage.DataDir, storage.WalDir, storage.BlockchainDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Ensure blockchain ledger exists
	if err := storage.EnsureBlockchainLedgerExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure blockchain ledger exists: %v", err)
	}

	return storage, nil
}

// EnsureBlockchainLedgerExists ensures the blockchain ledger file exists
func (s *DataStorage) EnsureBlockchainLedgerExists() error {
	if _, err := os.Stat(s.BlockchainFile); os.IsNotExist(err) {
		log.Printf("Creating new blockchain ledger file at %s", s.BlockchainFile)

		// Create an empty ledger file with initial structure
		initialData := BlockchainLedger{
			Blocks:      []Block{},
			LastUpdated: time.Now().Format(time.RFC3339),
			BlockHeight: 0,
		}

		data, err := json.MarshalIndent(initialData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal initial blockchain data: %v", err)
		}

		if err := os.WriteFile(s.BlockchainFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write blockchain ledger file: %v", err)
		}
	}

	return nil
}

// GetBlockchainLedger retrieves the blockchain ledger data
func (s *DataStorage) GetBlockchainLedger() (*BlockchainLedger, error) {
	if err := s.EnsureBlockchainLedgerExists(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.BlockchainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read blockchain ledger: %v", err)
	}

	var ledger BlockchainLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blockchain ledger: %v", err)
	}

	return &ledger, nil
}

// AddTransactionToBlockchain adds a transaction to the blockchain ledger
func (s *DataStorage) AddTransactionToBlockchain(txData map[string]interface{}, txHash string) error {
	// Add to local blockchain ledger
	if err := s.EnsureBlockchainLedgerExists(); err != nil {
		return fmt.Errorf("could not ensure blockchain ledger exists: %v", err)
	}

	// Read the current ledger data
	ledger, err := s.GetBlockchainLedger()
	if err != nil {
		return fmt.Errorf("could not read blockchain ledger: %v", err)
	}

	// Get the current block height and increment it
	currentHeight := ledger.BlockHeight
	newHeight := currentHeight + 1

	// Create a new block
	var previousBlockHash string
	if len(ledger.Blocks) > 0 {
		previousBlockHash = ledger.Blocks[len(ledger.Blocks)-1].TxHash
	}

	newBlock := Block{
		BlockHeight:       newHeight,
		TxHash:            txHash,
		TxData:            txData,
		Timestamp:         time.Now().Format(time.RFC3339),
		PreviousBlockHash: previousBlockHash,
	}

	// Add the new block to the ledger
	ledger.Blocks = append(ledger.Blocks, newBlock)
	ledger.BlockHeight = newHeight
	ledger.LastUpdated = time.Now().Format(time.RFC3339)

	// Write the updated ledger data back to the file
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated blockchain data: %v", err)
	}

	if err := os.WriteFile(s.BlockchainFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write updated blockchain ledger: %v", err)
	}

	// Update Supabase with the transaction hash
	// First, ensure the blockchain_ledger table exists
	_, err = s.Supabase.Select("blockchain_ledger", "count", map[string]interface{}{"limit": 1})
	if err != nil {
		// Create the table if it doesn't exist
		createTableSQL := `
			CREATE TABLE IF NOT EXISTS blockchain_ledger (
				id SERIAL PRIMARY KEY,
				tx_hash TEXT NOT NULL,
				tx_data JSONB NOT NULL,
				block_height INTEGER NOT NULL,
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL
			);
		`
		// Instead of using RPC, we'll just log the SQL that needs to be executed
		log.Printf("Please execute the following SQL in your Supabase database:\n%s", createTableSQL)
	}

	// Add transaction to blockchain_ledger table
	_, err = s.Supabase.Insert("blockchain_ledger", map[string]interface{}{
		"tx_hash":      txHash,
		"tx_data":      txData,
		"block_height": newHeight,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("Warning: Failed to add transaction to blockchain_ledger table: %v", err)
	} else {
		log.Printf("Added transaction to blockchain_ledger table: %s", txHash)
	}

	// Update the corresponding record
	if drugID, ok := txData["drug_id"].(string); ok {
		updateData := map[string]interface{}{
			"blockchain_tx_id": txHash,
		}
		_, err = s.Supabase.Update("drugs", drugID, updateData)
		if err != nil {
			log.Printf("Warning: Failed to update blockchain_tx_id for drug %s: %v", drugID, err)
		} else {
			log.Printf("Updated blockchain_tx_id for drug %s", drugID)

			// Only save to data_records after successful update of blockchain_tx_id
			recordFile := filepath.Join(s.DataDir, fmt.Sprintf("drug_%s.json", drugID))

			// Update the txData with the final blockchain_tx_id
			txData["blockchain_tx_id"] = txHash

			recordData, err := json.MarshalIndent(txData, "", "  ")
			if err != nil {
				log.Printf("Warning: Failed to marshal drug record: %v", err)
			} else {
				if err := os.WriteFile(recordFile, recordData, 0644); err != nil {
					log.Printf("Warning: Failed to write drug record to file: %v", err)
				} else {
					log.Printf("Saved drug record to %s", recordFile)
				}
			}

			// Update manufacturer and common ledgers
			if manufacturerID, ok := txData["manufacturer"].(string); ok {
				// Initialize ledger storage
				ledgerStorage, err := NewLedgerStorage()
				if err != nil {
					log.Printf("Warning: Failed to initialize ledger storage: %v", err)
				} else {
					// Get manufacturer ledger
					manufacturerLedger, err := ledgerStorage.GetManufacturerLedger(manufacturerID)
					if err != nil {
						log.Printf("Warning: Failed to get manufacturer ledger: %v", err)
					} else {
						// Create drug record in manufacturer ledger
						timestamp := time.Now().Format(time.RFC3339)
						drugRecord := models.DrugRecord{
							DrugID:        drugID,
							Status:        "created",
							CreatedAt:     timestamp,
							CurrentStatus: "created",
							History: []models.Status{
								{
									Status:    "created",
									Timestamp: timestamp,
									Details:   "Drug created",
								},
							},
						}

						// Add drug to manufacturer ledger
						manufacturerLedger.Drugs = append(manufacturerLedger.Drugs, drugRecord)
						manufacturerLedger.LastUpdated = timestamp

						// Save manufacturer ledger
						if err := ledgerStorage.SaveManufacturerLedger(manufacturerLedger); err != nil {
							log.Printf("Warning: Failed to save manufacturer ledger: %v", err)
						} else {
							log.Printf("Updated manufacturer ledger for %s", manufacturerID)
						}

						// Get common ledger
						commonLedger, err := ledgerStorage.GetCommonLedger()
						if err != nil {
							log.Printf("Warning: Failed to get common ledger: %v", err)
						} else {
							// Create drug record in common ledger
							commonDrugRecord := models.CommonDrugRecord{
								DrugID:         drugID,
								ManufacturerID: manufacturerID,
								Status:         "created",
								CreatedAt:      timestamp,
								CurrentStatus:  "created",
								History: []models.Status{
									{
										Status:    "created",
										Timestamp: timestamp,
										Details:   "Drug created",
									},
								},
								VerificationHash: txHash,
							}

							// Add drug to common ledger
							commonLedger.Drugs = append(commonLedger.Drugs, commonDrugRecord)
							commonLedger.LastUpdated = timestamp

							// Save common ledger
							if err := ledgerStorage.SaveCommonLedger(commonLedger); err != nil {
								log.Printf("Warning: Failed to save common ledger: %v", err)
							} else {
								log.Printf("Updated common ledger with drug %s", drugID)
							}
						}
					}
				}
			}
		}
	} else if shipmentID, ok := txData["shipment_id"].(string); ok {
		updateData := map[string]interface{}{
			"blockchain_tx_id": txHash,
		}
		_, err = s.Supabase.Update("shipments", shipmentID, updateData)
		if err != nil {
			log.Printf("Warning: Failed to update blockchain_tx_id for shipment %s: %v", shipmentID, err)
		} else {
			log.Printf("Updated blockchain_tx_id for shipment %s", shipmentID)

			// Only save to data_records after successful update of blockchain_tx_id
			recordFile := filepath.Join(s.DataDir, fmt.Sprintf("shipment_%s.json", shipmentID))

			// Update the txData with the final blockchain_tx_id
			txData["blockchain_tx_id"] = txHash

			recordData, err := json.MarshalIndent(txData, "", "  ")
			if err != nil {
				log.Printf("Warning: Failed to marshal shipment record: %v", err)
			} else {
				if err := os.WriteFile(recordFile, recordData, 0644); err != nil {
					log.Printf("Warning: Failed to write shipment record to file: %v", err)
				} else {
					log.Printf("Saved shipment record to %s", recordFile)
				}
			}
		}
	}

	log.Printf("Added transaction %s to blockchain ledger at block height %d", txHash, newHeight)
	return nil
}

// RunConsistencyCheck runs consistency checks and returns a detailed report
func (s *DataStorage) RunConsistencyCheck() ConsistencyCheckResult {
	log.Println("Starting blockchain-database consistency check")

	// Check database connectivity
	_, err := s.Supabase.Select("drugs", "count", map[string]interface{}{"limit": 1})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return ConsistencyCheckResult{
			Status:    "failed",
			Error:     "Database connection failed",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// Run the detailed consistency check
	isConsistent, err := s.checkConsistency()
	if err != nil {
		log.Printf("Consistency check error: %v", err)
		return ConsistencyCheckResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	if isConsistent {
		log.Println("Blockchain-database consistency check passed")
		return ConsistencyCheckResult{
			Status:    "consistent",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	} else {
		log.Println("Blockchain-database consistency check failed")
		return ConsistencyCheckResult{
			Status:         "inconsistent",
			Timestamp:      time.Now().Format(time.RFC3339),
			Recommendation: "Run manual verification and repair process",
		}
	}
}

// checkConsistency performs the actual consistency check
func (s *DataStorage) checkConsistency() (bool, error) {
	// This is a simplified implementation
	// In a real-world scenario, this would perform more thorough checks

	// Get the blockchain ledger
	ledger, err := s.GetBlockchainLedger()
	if err != nil {
		return false, fmt.Errorf("failed to get blockchain ledger: %v", err)
	}

	// Check if the blockchain is empty
	if len(ledger.Blocks) == 0 {
		// Empty blockchain is considered consistent
		return true, nil
	}

	// Check block sequence integrity
	for i := 1; i < len(ledger.Blocks); i++ {
		currentBlock := ledger.Blocks[i]
		previousBlock := ledger.Blocks[i-1]

		// Check block height sequence
		if currentBlock.BlockHeight != previousBlock.BlockHeight+1 {
			return false, fmt.Errorf("block height sequence broken at block %d", i)
		}

		// Check previous block hash reference
		if currentBlock.PreviousBlockHash != previousBlock.TxHash {
			return false, fmt.Errorf("previous block hash mismatch at block %d", i)
		}
	}

	return true, nil
}

// WriteFile writes data to a file in the specified path
func (s *DataStorage) WriteFile(filePath string, data []byte) error {
	// Ensure the directory exists
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return nil
}
