package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ankit/blockchain_ledger/models"
	"github.com/ankit/blockchain_ledger/storage"
)

// LedgerManager handles the synchronization between manufacturer ledgers and the common ledger
type LedgerManager struct {
	storage *storage.LedgerStorage
}

// NewLedgerManager creates a new ledger manager
func NewLedgerManager(storage *storage.LedgerStorage) *LedgerManager {
	return &LedgerManager{
		storage: storage,
	}
}

// SyncManufacturerLedger synchronizes a manufacturer's ledger with the common ledger
func (lm *LedgerManager) SyncManufacturerLedger(manufacturerID string) error {
	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(manufacturerID)
	if err != nil {
		return fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Update drug records in common ledger
	for _, drug := range manufacturerLedger.Drugs {
		found := false
		for i, commonDrug := range commonLedger.Drugs {
			if commonDrug.DrugID == drug.DrugID {
				// Update existing drug record
				commonLedger.Drugs[i] = models.CommonDrugRecord{
					DrugID:         drug.DrugID,
					ManufacturerID: manufacturerID,
					Status:         drug.Status,
					CreatedAt:      drug.CreatedAt,
					CurrentStatus:  drug.CurrentStatus,
					History:        drug.History,
				}
				found = true
				break
			}
		}
		if !found {
			// Add new drug record
			commonLedger.Drugs = append(commonLedger.Drugs, models.CommonDrugRecord{
				DrugID:         drug.DrugID,
				ManufacturerID: manufacturerID,
				Status:         drug.Status,
				CreatedAt:      drug.CreatedAt,
				CurrentStatus:  drug.CurrentStatus,
				History:        drug.History,
			})
		}
	}

	// Update shipment records in common ledger
	for _, shipment := range manufacturerLedger.Shipments {
		found := false
		for i, commonShipment := range commonLedger.Shipments {
			if commonShipment.ShipmentID == shipment.ShipmentID {
				// Update existing shipment record
				commonLedger.Shipments[i] = models.CommonShipmentRecord{
					ShipmentID:     shipment.ShipmentID,
					DrugID:         shipment.DrugID,
					ManufacturerID: manufacturerID,
					Status:         shipment.Status,
					CreatedAt:      shipment.CreatedAt,
					CurrentStatus:  shipment.CurrentStatus,
					History:        shipment.History,
				}
				found = true
				break
			}
		}
		if !found {
			// Add new shipment record
			commonLedger.Shipments = append(commonLedger.Shipments, models.CommonShipmentRecord{
				ShipmentID:     shipment.ShipmentID,
				DrugID:         shipment.DrugID,
				ManufacturerID: manufacturerID,
				Status:         shipment.Status,
				CreatedAt:      shipment.CreatedAt,
				CurrentStatus:  shipment.CurrentStatus,
				History:        shipment.History,
			})
		}
	}

	// Save updated common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	return nil
}

// SyncAllLedgers synchronizes all manufacturer ledgers with the common ledger
func (lm *LedgerManager) SyncAllLedgers() error {
	// Get list of all manufacturer IDs
	manufacturerIDs, err := lm.storage.ListManufacturerLedgers()
	if err != nil {
		return fmt.Errorf("failed to list manufacturer ledgers: %v", err)
	}

	// Sync each manufacturer's ledger
	for _, manufacturerID := range manufacturerIDs {
		if err := lm.SyncManufacturerLedger(manufacturerID); err != nil {
			return fmt.Errorf("failed to sync manufacturer ledger %s: %v", manufacturerID, err)
		}
	}

	return nil
}

// GetManufacturerLedger retrieves a manufacturer's ledger
func (lm *LedgerManager) GetManufacturerLedger(manufacturerID string) (*models.ManufacturerLedger, error) {
	return lm.storage.GetManufacturerLedger(manufacturerID)
}

// GetCommonLedger retrieves the common ledger
func (lm *LedgerManager) GetCommonLedger() (*models.CommonLedger, error) {
	return lm.storage.GetCommonLedger()
}

// UpdateDrugStatus updates a drug's status in both manufacturer and common ledgers
func (lm *LedgerManager) UpdateDrugStatus(manufacturerID, drugID, status string, details string) error {
	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(manufacturerID)
	if err != nil {
		return fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Update drug status in manufacturer ledger
	for i, drug := range manufacturerLedger.Drugs {
		if drug.DrugID == drugID {
			manufacturerLedger.Drugs[i].Status = status
			manufacturerLedger.Drugs[i].CurrentStatus = status
			manufacturerLedger.Drugs[i].History = append(manufacturerLedger.Drugs[i].History, models.Status{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			break
		}
	}

	// Save manufacturer ledger
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Update drug status in common ledger
	for i, drug := range commonLedger.Drugs {
		if drug.DrugID == drugID {
			commonLedger.Drugs[i].Status = status
			commonLedger.Drugs[i].CurrentStatus = status
			commonLedger.Drugs[i].History = append(commonLedger.Drugs[i].History, models.Status{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			break
		}
	}

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	return nil
}

// UpdateShipmentStatus updates a shipment's status in both manufacturer and common ledgers
func (lm *LedgerManager) UpdateShipmentStatus(manufacturerID, shipmentID, status string, details string) error {
	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(manufacturerID)
	if err != nil {
		return fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Update shipment status in manufacturer ledger
	for i, shipment := range manufacturerLedger.Shipments {
		if shipment.ShipmentID == shipmentID {
			manufacturerLedger.Shipments[i].Status = status
			manufacturerLedger.Shipments[i].CurrentStatus = status
			manufacturerLedger.Shipments[i].History = append(manufacturerLedger.Shipments[i].History, models.Status{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			break
		}
	}

	// Save manufacturer ledger
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Update shipment status in common ledger
	for i, shipment := range commonLedger.Shipments {
		if shipment.ShipmentID == shipmentID {
			commonLedger.Shipments[i].Status = status
			commonLedger.Shipments[i].CurrentStatus = status
			commonLedger.Shipments[i].History = append(commonLedger.Shipments[i].History, models.Status{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			break
		}
	}

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	return nil
}

// CreateDrug creates a new drug in both manufacturer and common ledgers
func (lm *LedgerManager) CreateDrug(params *models.CreateDrugParams) (string, error) {
	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(params.ManufacturerID)
	if err != nil {
		return "", fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Create drug record
	drugRecord := models.DrugRecord{
		DrugID:        params.DrugID,
		Status:        "created",
		CreatedAt:     time.Now().Format(time.RFC3339),
		CurrentStatus: "created",
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   "Drug created",
			},
		},
	}

	// Add drug to manufacturer ledger
	manufacturerLedger.Drugs = append(manufacturerLedger.Drugs, drugRecord)
	manufacturerLedger.LastUpdated = time.Now().Format(time.RFC3339)

	// Save manufacturer ledger
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return "", fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return "", fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Create common drug record
	commonDrugRecord := models.CommonDrugRecord{
		DrugID:         params.DrugID,
		ManufacturerID: params.ManufacturerID,
		Status:         "created",
		CreatedAt:      time.Now().Format(time.RFC3339),
		CurrentStatus:  "created",
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   "Drug created",
			},
		},
	}

	// Add drug to common ledger
	commonLedger.Drugs = append(commonLedger.Drugs, commonDrugRecord)
	commonLedger.LastUpdated = time.Now().Format(time.RFC3339)

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return "", fmt.Errorf("failed to save common ledger: %v", err)
	}

	return params.DrugID, nil
}

// CreateShipment creates a new shipment in both manufacturer and common ledgers
func (lm *LedgerManager) CreateShipment(params *models.CreateShipmentParams) (string, error) {
	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(params.ManufacturerID)
	if err != nil {
		return "", fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Create shipment record
	shipmentRecord := models.ShipmentRecord{
		ShipmentID:    params.ShipmentID,
		DrugID:        params.DrugID,
		Status:        "created",
		CreatedAt:     time.Now().Format(time.RFC3339),
		CurrentStatus: "created",
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   "Shipment created",
			},
		},
	}

	// Add shipment to manufacturer ledger
	manufacturerLedger.Shipments = append(manufacturerLedger.Shipments, shipmentRecord)
	manufacturerLedger.LastUpdated = time.Now().Format(time.RFC3339)

	// Save manufacturer ledger
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return "", fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return "", fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Create common shipment record
	commonShipmentRecord := models.CommonShipmentRecord{
		ShipmentID:     params.ShipmentID,
		DrugID:         params.DrugID,
		ManufacturerID: params.ManufacturerID,
		DistributorID:  params.DistributorID,
		Status:         "created",
		CreatedAt:      time.Now().Format(time.RFC3339),
		CurrentStatus:  "created",
		History: []models.Status{
			{
				Status:    "created",
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   "Shipment created",
			},
		},
	}

	// Add shipment to common ledger
	commonLedger.Shipments = append(commonLedger.Shipments, commonShipmentRecord)
	commonLedger.LastUpdated = time.Now().Format(time.RFC3339)

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return "", fmt.Errorf("failed to save common ledger: %v", err)
	}

	return params.ShipmentID, nil
}

// RevertDrug reverts a drug in both manufacturer and common ledgers
func (lm *LedgerManager) RevertDrug(params *models.RevertDrugParams) error {
	// Get manufacturer ledger
	manufacturerLedger, err := lm.storage.GetManufacturerLedger(params.ManufacturerID)
	if err != nil {
		return fmt.Errorf("failed to get manufacturer ledger: %v", err)
	}

	// Update drug status in manufacturer ledger
	for i, drug := range manufacturerLedger.Drugs {
		if drug.DrugID == params.DrugID {
			manufacturerLedger.Drugs[i].Status = "reverted"
			manufacturerLedger.Drugs[i].CurrentStatus = "reverted"
			manufacturerLedger.Drugs[i].RevertedAt = time.Now().Format(time.RFC3339)
			manufacturerLedger.Drugs[i].History = append(manufacturerLedger.Drugs[i].History, models.Status{
				Status:    "reverted",
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   params.Reason,
			})
			break
		}
	}

	// Save manufacturer ledger
	if err := lm.storage.SaveManufacturerLedger(manufacturerLedger); err != nil {
		return fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Update drug status in common ledger
	for i, drug := range commonLedger.Drugs {
		if drug.DrugID == params.DrugID {
			commonLedger.Drugs[i].Status = "reverted"
			commonLedger.Drugs[i].CurrentStatus = "reverted"
			commonLedger.Drugs[i].History = append(commonLedger.Drugs[i].History, models.Status{
				Status:    "reverted",
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   params.Reason,
			})
			break
		}
	}

	// Save common ledger
	if err := lm.storage.SaveCommonLedger(commonLedger); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	return nil
}

// VerifyDrug verifies a drug's authenticity
func (lm *LedgerManager) VerifyDrug(drugID string) (bool, error) {
	// Get common ledger
	commonLedger, err := lm.storage.GetCommonLedger()
	if err != nil {
		return false, fmt.Errorf("failed to get common ledger: %v", err)
	}

	// Find drug in common ledger
	for _, drug := range commonLedger.Drugs {
		if drug.DrugID == drugID {
			return true, nil
		}
	}

	return false, nil
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
func (lm *LedgerManager) generateVerificationHash(drugID string, manufacturerID string) string {
	// Create a unique string combining drug ID, manufacturer ID, and timestamp
	data := fmt.Sprintf("%s:%s:%s", drugID, manufacturerID, time.Now().Format(time.RFC3339Nano))

	// Generate SHA-256 hash
	hash := sha256.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}
