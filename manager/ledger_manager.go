package manager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ankit/blockchain_ledger/blockchain"
	"github.com/ankit/blockchain_ledger/models"
	"github.com/ankit/blockchain_ledger/storage"
)

// LedgerManager implements the models.LedgerManager interface
type LedgerManager struct {
	storage    *storage.LedgerStorage
	blockchain *blockchain.BlockchainService
}

// NewLedgerManager creates a new ledger manager
func NewLedgerManager(storage *storage.LedgerStorage, blockchain *blockchain.BlockchainService) *LedgerManager {
	return &LedgerManager{
		storage:    storage,
		blockchain: blockchain,
	}
}

// CreateDrug creates a new drug in the manufacturer and common ledgers
func (lm *LedgerManager) CreateDrug(params *models.CreateDrugParams) (string, error) {
	// Get current timestamp
	now := time.Now()
	timestamp := now.Format(time.RFC3339)

	// Generate verification hash
	verificationHash := lm.generateVerificationHash(params.DrugID, params.ManufacturerID, timestamp)

	// Create blockchain transaction
	txData := map[string]interface{}{
		"drug_id":           params.DrugID,
		"manufacturer_id":   params.ManufacturerID,
		"name":              params.Name,
		"description":       params.Description,
		"verification_hash": verificationHash,
		"created_at":        timestamp,
	}
	txHash, err := lm.blockchain.CreateTransaction("drug_create", txData)
	if err != nil {
		return "", fmt.Errorf("failed to create blockchain transaction: %v", err)
	}

	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(params.ManufacturerID)
	if err != nil {
		return "", fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Create drug record in manufacturer ledger
	drugRecord := models.DrugRecord{
		DrugID:        params.DrugID,
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
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return "", fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return "", fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Create drug record in common ledger
	commonDrugRecord := models.CommonDrugRecord{
		DrugID:           params.DrugID,
		ManufacturerID:   params.ManufacturerID,
		Status:           "created",
		CreatedAt:        timestamp,
		CurrentStatus:    "created",
		VerificationHash: verificationHash,
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: timestamp,
				Details:   "Drug created",
			},
		},
	}

	// Add drug to common ledger
	commonLedger.Drugs = append(commonLedger.Drugs, commonDrugRecord)
	commonLedger.LastUpdated = timestamp

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return "", fmt.Errorf("failed to save common ledger: %v", err)
	}

	// Insert drug into database
	drug := &models.Drug{
		ID:               params.DrugID,
		ManufacturerID:   params.ManufacturerID,
		Name:             params.Name,
		Description:      params.Description,
		Status:           "created",
		VerificationHash: verificationHash,
		BlockchainTxID:   txHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := lm.storage.InsertDrug(drug); err != nil {
		return "", fmt.Errorf("failed to insert drug into database: %v", err)
	}

	// Insert drug status update into database
	drugStatusUpdate := &models.DrugStatusUpdate{
		DrugID:         params.DrugID,
		Status:         "created",
		Location:       params.Location,
		UpdatedBy:      params.UserID,
		BlockchainTxID: txHash,
		Timestamp:      now,
	}
	if err := lm.storage.InsertDrugStatusUpdate(drugStatusUpdate); err != nil {
		return "", fmt.Errorf("failed to insert drug status update into database: %v", err)
	}

	return verificationHash, nil
}

// CreateShipment creates a new shipment in the manufacturer and common ledgers
func (lm *LedgerManager) CreateShipment(params *models.CreateShipmentParams) (string, error) {
	// Get current timestamp
	now := time.Now()
	timestamp := now.Format(time.RFC3339)

	// Create blockchain transaction
	txData := map[string]interface{}{
		"shipment_id":     params.ShipmentID,
		"drug_id":         params.DrugID,
		"manufacturer_id": params.ManufacturerID,
		"distributor_id":  params.DistributorID,
		"created_at":      timestamp,
	}
	txHash, err := lm.blockchain.CreateTransaction("shipment_create", txData)
	if err != nil {
		return "", fmt.Errorf("failed to create blockchain transaction: %v", err)
	}

	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(params.ManufacturerID)
	if err != nil {
		return "", fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Create shipment record in manufacturer ledger
	shipmentRecord := models.ShipmentRecord{
		ShipmentID:    params.ShipmentID,
		DrugID:        params.DrugID,
		Status:        "created",
		CreatedAt:     timestamp,
		CurrentStatus: "created",
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: timestamp,
				Details:   "Shipment created",
			},
		},
	}

	// Add shipment to manufacturer ledger
	manufacturerLedger.Shipments = append(manufacturerLedger.Shipments, shipmentRecord)
	manufacturerLedger.LastUpdated = timestamp

	// Update drug status in manufacturer ledger
	for i, drug := range manufacturerLedger.Drugs {
		if drug.DrugID == params.DrugID {
			manufacturerLedger.Drugs[i].Status = "in_transit"
			manufacturerLedger.Drugs[i].CurrentStatus = "in_transit"
			manufacturerLedger.Drugs[i].History = append(manufacturerLedger.Drugs[i].History, models.Status{
				Status:    "in_transit",
				Timestamp: timestamp,
				Details:   fmt.Sprintf("Drug added to shipment %s", params.ShipmentID),
			})
			break
		}
	}

	// Save manufacturer ledger
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return "", fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return "", fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Create shipment record in common ledger
	commonShipmentRecord := models.CommonShipmentRecord{
		ShipmentID:     params.ShipmentID,
		DrugID:         params.DrugID,
		ManufacturerID: params.ManufacturerID,
		DistributorID:  params.DistributorID,
		Status:         "created",
		CreatedAt:      timestamp,
		CurrentStatus:  "created",
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: timestamp,
				Details:   "Shipment created",
			},
		},
	}

	// Add shipment to common ledger
	commonLedger.Shipments = append(commonLedger.Shipments, commonShipmentRecord)
	commonLedger.LastUpdated = timestamp

	// Update drug status in common ledger
	for i, drug := range commonLedger.Drugs {
		if drug.DrugID == params.DrugID {
			commonLedger.Drugs[i].Status = "in_transit"
			commonLedger.Drugs[i].CurrentStatus = "in_transit"
			commonLedger.Drugs[i].History = append(commonLedger.Drugs[i].History, models.Status{
				Status:    "in_transit",
				Timestamp: timestamp,
				Details:   fmt.Sprintf("Drug added to shipment %s", params.ShipmentID),
			})
			break
		}
	}

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return "", fmt.Errorf("failed to save common ledger: %v", err)
	}

	// Insert shipment into database
	shipment := &models.Shipment{
		ID:             params.ShipmentID,
		DrugID:         params.DrugID,
		ManufacturerID: params.ManufacturerID,
		DistributorID:  params.DistributorID,
		Status:         "created",
		BlockchainTxID: txHash,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := lm.storage.InsertShipment(shipment); err != nil {
		return "", fmt.Errorf("failed to insert shipment into database: %v", err)
	}

	// Insert shipment status update into database
	shipmentStatusUpdate := &models.ShipmentStatusUpdate{
		ShipmentID:     params.ShipmentID,
		Status:         "created",
		Location:       params.Location,
		UpdatedBy:      params.UserID,
		BlockchainTxID: txHash,
		Timestamp:      now,
	}
	if err := lm.storage.InsertShipmentStatusUpdate(shipmentStatusUpdate); err != nil {
		return "", fmt.Errorf("failed to insert shipment status update into database: %v", err)
	}

	// Update drug status in database
	drugStatusTxHash, err := lm.blockchain.RecordDrugStatusUpdate(params.DrugID, "in_transit", params.UserID, now)
	if err != nil {
		return "", fmt.Errorf("failed to record drug status update in blockchain: %v", err)
	}

	// Get drug from database
	drug, err := lm.storage.GetDrug(params.DrugID)
	if err != nil {
		return "", fmt.Errorf("failed to get drug from database: %v", err)
	}

	// Update drug status
	drug.Status = "in_transit"
	drug.BlockchainTxID = drugStatusTxHash
	drug.UpdatedAt = now

	// Update drug in database
	if err := lm.storage.UpdateDrug(drug); err != nil {
		return "", fmt.Errorf("failed to update drug in database: %v", err)
	}

	// Insert drug status update into database
	drugInTransitUpdate := &models.DrugStatusUpdate{
		DrugID:         params.DrugID,
		Status:         "in_transit",
		Location:       params.Location,
		UpdatedBy:      params.UserID,
		BlockchainTxID: drugStatusTxHash,
		Timestamp:      now,
	}
	if err := lm.storage.InsertDrugStatusUpdate(drugInTransitUpdate); err != nil {
		return "", fmt.Errorf("failed to insert drug status update into database: %v", err)
	}

	return txHash, nil
}

// UpdateShipmentStatus updates a shipment's status in the manufacturer and common ledgers
func (lm *LedgerManager) UpdateShipmentStatus(params *models.UpdateShipmentStatusParams) error {
	// Get current timestamp
	now := time.Now()
	timestamp := now.Format(time.RFC3339)

	// Create blockchain transaction
	txData := map[string]interface{}{
		"shipment_id": params.ShipmentID,
		"status":      params.Status,
		"updated_by":  params.UserID,
		"updated_at":  timestamp,
	}
	txHash, err := lm.blockchain.CreateTransaction("shipment_update", txData)
	if err != nil {
		return fmt.Errorf("failed to create blockchain transaction: %v", err)
	}

	// Get shipment from database
	shipment, err := lm.storage.GetShipment(params.ShipmentID)
	if err != nil {
		return fmt.Errorf("failed to get shipment from database: %v", err)
	}

	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(shipment.ManufacturerID)
	if err != nil {
		return fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Update shipment status in manufacturer ledger
	for i, s := range manufacturerLedger.Shipments {
		if s.ShipmentID == params.ShipmentID {
			manufacturerLedger.Shipments[i].Status = params.Status
			manufacturerLedger.Shipments[i].CurrentStatus = params.Status
			manufacturerLedger.Shipments[i].History = append(manufacturerLedger.Shipments[i].History, models.Status{
				Status:    params.Status,
				Timestamp: timestamp,
				Details:   fmt.Sprintf("Shipment status updated to %s", params.Status),
			})
			break
		}
	}

	// If shipment is delivered, update drug status
	if params.Status == "delivered" {
		// Update drug status in manufacturer ledger
		for i, drug := range manufacturerLedger.Drugs {
			if drug.DrugID == shipment.DrugID {
				manufacturerLedger.Drugs[i].Status = "delivered"
				manufacturerLedger.Drugs[i].CurrentStatus = "delivered"
				manufacturerLedger.Drugs[i].History = append(manufacturerLedger.Drugs[i].History, models.Status{
					Status:    "delivered",
					Timestamp: timestamp,
					Details:   fmt.Sprintf("Drug delivered via shipment %s", params.ShipmentID),
				})
				break
			}
		}
	}

	// Save manufacturer ledger
	manufacturerLedger.LastUpdated = timestamp
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Update shipment status in common ledger
	for i, s := range commonLedger.Shipments {
		if s.ShipmentID == params.ShipmentID {
			commonLedger.Shipments[i].Status = params.Status
			commonLedger.Shipments[i].CurrentStatus = params.Status
			commonLedger.Shipments[i].History = append(commonLedger.Shipments[i].History, models.Status{
				Status:    params.Status,
				Timestamp: timestamp,
				Details:   fmt.Sprintf("Shipment status updated to %s", params.Status),
			})
			break
		}
	}

	// If shipment is delivered, update drug status
	if params.Status == "delivered" {
		// Update drug status in common ledger
		for i, drug := range commonLedger.Drugs {
			if drug.DrugID == shipment.DrugID {
				commonLedger.Drugs[i].Status = "delivered"
				commonLedger.Drugs[i].CurrentStatus = "delivered"
				commonLedger.Drugs[i].History = append(commonLedger.Drugs[i].History, models.Status{
					Status:    "delivered",
					Timestamp: timestamp,
					Details:   fmt.Sprintf("Drug delivered via shipment %s", params.ShipmentID),
				})
				break
			}
		}
	}

	// Save common ledger
	commonLedger.LastUpdated = timestamp
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	// Update shipment in database
	shipment.Status = params.Status
	shipment.BlockchainTxID = txHash
	shipment.UpdatedAt = now

	if err := lm.storage.UpdateShipment(shipment); err != nil {
		return fmt.Errorf("failed to update shipment in database: %v", err)
	}

	// Insert shipment status update into database
	shipmentStatusUpdate := &models.ShipmentStatusUpdate{
		ShipmentID:     params.ShipmentID,
		Status:         params.Status,
		Location:       params.Location,
		UpdatedBy:      params.UserID,
		BlockchainTxID: txHash,
		Timestamp:      now,
	}
	if err := lm.storage.InsertShipmentStatusUpdate(shipmentStatusUpdate); err != nil {
		return fmt.Errorf("failed to insert shipment status update into database: %v", err)
	}

	// If shipment is delivered, update drug status
	if params.Status == "delivered" {
		// Create blockchain transaction for drug status update
		drugStatusTxHash, err := lm.blockchain.RecordDrugStatusUpdate(shipment.DrugID, "delivered", params.UserID, now)
		if err != nil {
			return fmt.Errorf("failed to record drug status update in blockchain: %v", err)
		}

		// Get drug from database
		drug, err := lm.storage.GetDrug(shipment.DrugID)
		if err != nil {
			return fmt.Errorf("failed to get drug from database: %v", err)
		}

		// Update drug status
		drug.Status = "delivered"
		drug.BlockchainTxID = drugStatusTxHash
		drug.UpdatedAt = now

		// Update drug in database
		if err := lm.storage.UpdateDrug(drug); err != nil {
			return fmt.Errorf("failed to update drug in database: %v", err)
		}

		// Insert drug status update into database
		drugStatusUpdate := &models.DrugStatusUpdate{
			DrugID:         shipment.DrugID,
			Status:         "delivered",
			Location:       params.Location,
			UpdatedBy:      params.UserID,
			BlockchainTxID: drugStatusTxHash,
			Timestamp:      now,
		}
		if err := lm.storage.InsertDrugStatusUpdate(drugStatusUpdate); err != nil {
			return fmt.Errorf("failed to insert drug status update into database: %v", err)
		}
	}

	return nil
}

// RevertDrug reverts a drug in the manufacturer and common ledgers
func (lm *LedgerManager) RevertDrug(params *models.RevertDrugParams) error {
	// Get current timestamp
	now := time.Now()
	timestamp := now.Format(time.RFC3339)

	// Create blockchain transaction
	txData := map[string]interface{}{
		"drug_id":    params.DrugID,
		"reason":     params.Reason,
		"updated_by": params.UserID,
		"updated_at": timestamp,
	}
	txHash, err := lm.blockchain.CreateTransaction("drug_revert", txData)
	if err != nil {
		return fmt.Errorf("failed to create blockchain transaction: %v", err)
	}

	// Get drug from database
	drug, err := lm.storage.GetDrug(params.DrugID)
	if err != nil {
		return fmt.Errorf("failed to get drug from database: %v", err)
	}

	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(drug.ManufacturerID)
	if err != nil {
		return fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Update drug status in manufacturer ledger
	for i, d := range manufacturerLedger.Drugs {
		if d.DrugID == params.DrugID {
			manufacturerLedger.Drugs[i].Status = "reverted"
			manufacturerLedger.Drugs[i].CurrentStatus = "reverted"
			manufacturerLedger.Drugs[i].RevertedAt = timestamp
			manufacturerLedger.Drugs[i].History = append(manufacturerLedger.Drugs[i].History, models.Status{
				Status:    "reverted",
				Timestamp: timestamp,
				Details:   fmt.Sprintf("Drug reverted: %s", params.Reason),
			})
			break
		}
	}

	// Save manufacturer ledger
	manufacturerLedger.LastUpdated = timestamp
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Update drug status in common ledger
	for i, d := range commonLedger.Drugs {
		if d.DrugID == params.DrugID {
			commonLedger.Drugs[i].Status = "reverted"
			commonLedger.Drugs[i].CurrentStatus = "reverted"
			commonLedger.Drugs[i].History = append(commonLedger.Drugs[i].History, models.Status{
				Status:    "reverted",
				Timestamp: timestamp,
				Details:   fmt.Sprintf("Drug reverted: %s", params.Reason),
			})
			break
		}
	}

	// Save common ledger
	commonLedger.LastUpdated = timestamp
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	// Update drug in database
	drug.Status = "reverted"
	drug.BlockchainTxID = txHash
	drug.UpdatedAt = now

	if err := lm.storage.UpdateDrug(drug); err != nil {
		return fmt.Errorf("failed to update drug in database: %v", err)
	}

	// Insert drug status update into database
	drugStatusUpdate := &models.DrugStatusUpdate{
		DrugID:         params.DrugID,
		Status:         "reverted",
		Location:       params.Location,
		UpdatedBy:      params.UserID,
		BlockchainTxID: txHash,
		Timestamp:      now,
	}
	if err := lm.storage.InsertDrugStatusUpdate(drugStatusUpdate); err != nil {
		return fmt.Errorf("failed to insert drug status update into database: %v", err)
	}

	return nil
}

// VerifyDrug verifies a drug's authenticity
func (lm *LedgerManager) VerifyDrug(drugID string) (bool, error) {
	// Get drug from database
	drug, err := lm.storage.GetDrug(drugID)
	if err != nil {
		return false, fmt.Errorf("failed to get drug from database: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return false, fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Find drug in common ledger
	var commonDrug *models.CommonDrugRecord
	for _, d := range commonLedger.Drugs {
		if d.DrugID == drugID {
			commonDrug = &d
			break
		}
	}

	if commonDrug == nil {
		return false, fmt.Errorf("drug not found in common ledger: %s", drugID)
	}

	// Verify drug's verification hash
	if drug.VerificationHash != commonDrug.VerificationHash {
		return false, nil
	}

	// Verify blockchain transaction
	return lm.blockchain.VerifyTransaction(drug.BlockchainTxID)
}

// GetDrugHistory retrieves a drug's status history
func (lm *LedgerManager) GetDrugHistory(drugID string) ([]models.DrugStatusUpdate, error) {
	return lm.storage.GetDrugStatusUpdates(drugID)
}

// GetShipmentHistory retrieves a shipment's status history
func (lm *LedgerManager) GetShipmentHistory(shipmentID string) ([]models.ShipmentStatusUpdate, error) {
	return lm.storage.GetShipmentStatusUpdates(shipmentID)
}

// generateVerificationHash generates a unique hash for drug verification
func (lm *LedgerManager) generateVerificationHash(drugID, manufacturerID, timestamp string) string {
	h := sha256.New()
	data := fmt.Sprintf("%s:%s:%s", drugID, manufacturerID, timestamp)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
