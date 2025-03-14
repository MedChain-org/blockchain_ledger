package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	Timestamp    string                 `json:"timestamp"`
	Data         map[string]interface{} `json:"data"`
	PreviousHash string                 `json:"previous_hash"`
	Hash         string                 `json:"hash"`
}

// NewTransaction creates a new blockchain transaction
func NewTransaction(data map[string]interface{}) *Transaction {
	tx := &Transaction{
		Timestamp:    time.Now().Format(time.RFC3339),
		Data:         data,
		PreviousHash: "",
	}
	tx.CalculateHash()
	return tx
}

// CalculateHash calculates the hash of the transaction
func (tx *Transaction) CalculateHash() {
	h := sha256.New()

	// Create a string representation of the transaction data
	dataJSON, _ := json.Marshal(tx.Data)
	data := fmt.Sprintf("%s%s%s", tx.Timestamp, string(dataJSON), tx.PreviousHash)

	h.Write([]byte(data))
	tx.Hash = hex.EncodeToString(h.Sum(nil))
}

// SetPreviousHash sets the previous hash for the transaction
func (tx *Transaction) SetPreviousHash(previousHash string) {
	tx.PreviousHash = previousHash
	tx.CalculateHash() // Recalculate hash after changing previous hash
}

// DrugTransaction represents a drug-related blockchain transaction
type DrugTransaction struct {
	Transaction
	DrugID string `json:"drug_id"`
}

// NewDrugTransaction creates a new drug transaction
func NewDrugTransaction(drugData map[string]interface{}) (*DrugTransaction, error) {
	// Validate required fields
	requiredFields := []string{"drug_id", "manufacturer", "name", "batch_number", "manufacture_date", "expiry_date"}
	for _, field := range requiredFields {
		if _, ok := drugData[field]; !ok {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
	}

	// Create transaction data with all required fields
	txData := map[string]interface{}{
		"drug_id":          drugData["drug_id"],
		"manufacturer_id":  drugData["manufacturer"],
		"name":             drugData["name"],
		"batch_number":     drugData["batch_number"],
		"manufacture_date": drugData["manufacture_date"],
		"expiry_date":      drugData["expiry_date"],
		"status":           "created",
		"created_at":       time.Now().Format(time.RFC3339),
	}

	// Create transaction
	tx := &DrugTransaction{
		Transaction: *NewTransaction(txData),
		DrugID:      drugData["drug_id"].(string),
	}

	return tx, nil
}

// ShipmentTransaction represents a shipment-related blockchain transaction
type ShipmentTransaction struct {
	Transaction
	ShipmentID string `json:"shipment_id"`
}

// NewShipmentTransaction creates a new shipment transaction
func NewShipmentTransaction(shipmentData map[string]interface{}) (*ShipmentTransaction, error) {
	// Validate required fields
	if _, ok := shipmentData["shipment_id"]; !ok {
		return nil, fmt.Errorf("missing required field: shipment_id")
	}

	// Create transaction
	tx := &ShipmentTransaction{
		Transaction: *NewTransaction(shipmentData),
		ShipmentID:  shipmentData["shipment_id"].(string),
	}

	return tx, nil
}

// GenerateTransactionHash generates a unique transaction hash
func GenerateTransactionHash(data map[string]interface{}) string {
	// Create a unique identifier
	uuid := uuid.New().String()

	// Combine with timestamp for uniqueness
	timestamp := time.Now().Format(time.RFC3339Nano)

	// Create a hash of the combined data
	h := sha256.New()
	dataJSON, _ := json.Marshal(data)
	h.Write([]byte(fmt.Sprintf("%s%s%s", uuid, timestamp, string(dataJSON))))

	return hex.EncodeToString(h.Sum(nil))
}

// ValidateTransaction validates a transaction hash against its data
func ValidateTransaction(txHash string, txData map[string]interface{}) bool {
	// Create a new transaction with the same data
	tx := NewTransaction(txData)

	// If the previous hash is included in the data, set it
	if prevHash, ok := txData["previous_hash"].(string); ok {
		tx.SetPreviousHash(prevHash)
	}

	// Compare the calculated hash with the provided hash
	return tx.Hash == txHash
}
