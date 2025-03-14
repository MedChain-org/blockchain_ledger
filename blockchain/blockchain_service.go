package blockchain

import (
	"fmt"
	"time"

	"github.com/ankit/blockchain_ledger/models"
	"github.com/ankit/blockchain_ledger/storage"
)

// BlockchainService implements the models.BlockchainService interface
type BlockchainService struct {
	dataStorage *storage.DataStorage
}

// NewBlockchainService creates a new blockchain service
func NewBlockchainService(dataStorage *storage.DataStorage) *BlockchainService {
	return &BlockchainService{
		dataStorage: dataStorage,
	}
}

// CreateTransaction creates a new blockchain transaction
func (bs *BlockchainService) CreateTransaction(txType string, data map[string]interface{}) (string, error) {
	// Add transaction type to data
	data["tx_type"] = txType
	data["timestamp"] = time.Now().Format(time.RFC3339)

	// Generate transaction hash
	txHash := GenerateTransactionHash(data)
	data["tx_hash"] = txHash

	// Add transaction to blockchain
	if err := bs.dataStorage.AddTransactionToBlockchain(data, txHash); err != nil {
		return "", fmt.Errorf("failed to add transaction to blockchain: %v", err)
	}

	return txHash, nil
}

// GetTransaction retrieves a transaction from the blockchain
func (bs *BlockchainService) GetTransaction(txID string) (*models.BlockchainTransaction, error) {
	// Get blockchain ledger
	ledger, err := bs.dataStorage.GetBlockchainLedger()
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain ledger: %v", err)
	}

	// Find transaction with matching hash
	for _, block := range ledger.Blocks {
		if block.TxHash == txID {
			// Convert block data to BlockchainTransaction
			txData, ok := block.TxData.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid transaction data format")
			}

			txType, _ := txData["tx_type"].(string)
			timestamp, _ := time.Parse(time.RFC3339, block.Timestamp)

			return &models.BlockchainTransaction{
				TransactionID: block.TxHash,
				Type:          txType,
				Data:          txData,
				Timestamp:     timestamp,
			}, nil
		}
	}

	return nil, fmt.Errorf("transaction not found: %s", txID)
}

// VerifyTransaction verifies a transaction in the blockchain
func (bs *BlockchainService) VerifyTransaction(txID string) (bool, error) {
	// Get blockchain ledger
	ledger, err := bs.dataStorage.GetBlockchainLedger()
	if err != nil {
		return false, fmt.Errorf("failed to get blockchain ledger: %v", err)
	}

	// Find transaction with matching hash
	for _, block := range ledger.Blocks {
		if block.TxHash == txID {
			// Convert block data to map
			txData, ok := block.TxData.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("invalid transaction data format")
			}

			// Validate transaction
			return ValidateTransaction(txID, txData), nil
		}
	}

	return false, fmt.Errorf("transaction not found: %s", txID)
}

// RecordDrugCreation records a drug creation in the blockchain
func (bs *BlockchainService) RecordDrugCreation(drugID, manufacturerID string, timestamp time.Time) (string, error) {
	data := map[string]interface{}{
		"drug_id":         drugID,
		"manufacturer_id": manufacturerID,
		"status":          "created",
		"created_at":      timestamp.Format(time.RFC3339),
	}

	return bs.CreateTransaction("drug_create", data)
}

// RecordDrugStatusUpdate records a drug status update in the blockchain
func (bs *BlockchainService) RecordDrugStatusUpdate(drugID, status, updatedBy string, timestamp time.Time) (string, error) {
	data := map[string]interface{}{
		"drug_id":    drugID,
		"status":     status,
		"updated_by": updatedBy,
		"updated_at": timestamp.Format(time.RFC3339),
	}

	return bs.CreateTransaction("drug_update", data)
}

// RecordShipmentCreation records a shipment creation in the blockchain
func (bs *BlockchainService) RecordShipmentCreation(shipmentID, drugID, manufacturerID, distributorID string, timestamp time.Time) (string, error) {
	data := map[string]interface{}{
		"shipment_id":     shipmentID,
		"drug_id":         drugID,
		"manufacturer_id": manufacturerID,
		"distributor_id":  distributorID,
		"status":          "created",
		"created_at":      timestamp.Format(time.RFC3339),
	}

	return bs.CreateTransaction("shipment_create", data)
}

// RecordShipmentStatusUpdate records a shipment status update in the blockchain
func (bs *BlockchainService) RecordShipmentStatusUpdate(shipmentID, status, updatedBy string, timestamp time.Time) (string, error) {
	data := map[string]interface{}{
		"shipment_id": shipmentID,
		"status":      status,
		"updated_by":  updatedBy,
		"updated_at":  timestamp.Format(time.RFC3339),
	}

	return bs.CreateTransaction("shipment_update", data)
}
