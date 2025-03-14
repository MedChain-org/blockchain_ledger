package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ankit/blockchain_ledger/models"
	"github.com/ankit/blockchain_ledger/supabase"
)

// LedgerStorage handles the storage of manufacturer and common ledgers
type LedgerStorage struct {
	ManufacturerLedgersDir string
	CommonLedgerPath       string
	Supabase               *supabase.Client
}

// NewLedgerStorage creates a new ledger storage instance
func NewLedgerStorage() (*LedgerStorage, error) {
	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient(true) // Use service key
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %v", err)
	}

	ls := &LedgerStorage{
		ManufacturerLedgersDir: "blockchain_data/manufacturer_ledgers",
		CommonLedgerPath:       "blockchain_data/common_ledger.json",
		Supabase:               supabaseClient,
	}

	// Create manufacturer ledgers directory if it doesn't exist
	if err := os.MkdirAll(ls.ManufacturerLedgersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create manufacturer ledgers directory: %v", err)
	}

	// Create common ledger file if it doesn't exist
	if _, err := os.Stat(ls.CommonLedgerPath); os.IsNotExist(err) {
		commonLedger := models.NewCommonLedger()
		data, err := commonLedger.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal initial common ledger: %v", err)
		}
		if err := os.WriteFile(ls.CommonLedgerPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to create common ledger file: %v", err)
		}
	}

	return ls, nil
}

// GetManufacturerLedger loads a manufacturer's ledger from disk
func (ls *LedgerStorage) GetManufacturerLedger(manufacturerID string) (*models.ManufacturerLedger, error) {
	ledgerPath := filepath.Join(ls.ManufacturerLedgersDir, fmt.Sprintf("%s.json", manufacturerID))

	// Check if the ledger file exists
	if _, err := os.Stat(ledgerPath); os.IsNotExist(err) {
		// Create a new ledger if it doesn't exist
		newLedger := models.NewManufacturerLedger(manufacturerID)
		if err := ls.SaveManufacturerLedger(newLedger); err != nil {
			return nil, fmt.Errorf("failed to create new manufacturer ledger: %v", err)
		}
		return newLedger, nil
	}

	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manufacturer ledger: %v", err)
	}

	var ledger models.ManufacturerLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manufacturer ledger: %v", err)
	}

	return &ledger, nil
}

// SaveManufacturerLedger saves a manufacturer's ledger to disk
func (ls *LedgerStorage) SaveManufacturerLedger(ledger *models.ManufacturerLedger) error {
	data, err := ledger.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal manufacturer ledger: %v", err)
	}

	ledgerPath := filepath.Join(ls.ManufacturerLedgersDir, fmt.Sprintf("%s.json", ledger.ManufacturerID))
	if err := os.WriteFile(ledgerPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save manufacturer ledger: %v", err)
	}

	return nil
}

// GetCommonLedger loads the common ledger from disk
func (ls *LedgerStorage) GetCommonLedger() (*models.CommonLedger, error) {
	data, err := os.ReadFile(ls.CommonLedgerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read common ledger: %v", err)
	}

	var ledger models.CommonLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshal common ledger: %v", err)
	}

	return &ledger, nil
}

// SaveCommonLedger saves the common ledger to disk
func (ls *LedgerStorage) SaveCommonLedger(ledger *models.CommonLedger) error {
	data, err := ledger.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal common ledger: %v", err)
	}

	if err := os.WriteFile(ls.CommonLedgerPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save common ledger: %v", err)
	}

	return nil
}

// ListManufacturerLedgers returns a list of all manufacturer IDs that have ledgers
func (ls *LedgerStorage) ListManufacturerLedgers() ([]string, error) {
	files, err := os.ReadDir(ls.ManufacturerLedgersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read manufacturer ledgers directory: %v", err)
	}

	var manufacturerIDs []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			manufacturerIDs = append(manufacturerIDs, file.Name()[:len(file.Name())-5])
		}
	}

	return manufacturerIDs, nil
}

// DeleteManufacturerLedger deletes a manufacturer's ledger
func (ls *LedgerStorage) DeleteManufacturerLedger(manufacturerID string) error {
	ledgerPath := filepath.Join(ls.ManufacturerLedgersDir, fmt.Sprintf("%s.json", manufacturerID))
	if err := os.Remove(ledgerPath); err != nil {
		return fmt.Errorf("failed to delete manufacturer ledger: %v", err)
	}
	return nil
}

// Database operations

// InsertDrug inserts a drug record into the database
func (ls *LedgerStorage) InsertDrug(drug *models.Drug) error {
	_, err := ls.Supabase.Insert("drugs", drug)
	return err
}

// UpdateDrug updates a drug record in the database
func (ls *LedgerStorage) UpdateDrug(drug *models.Drug) error {
	// Convert drug to map for update
	drugData, err := json.Marshal(drug)
	if err != nil {
		return fmt.Errorf("failed to marshal drug: %v", err)
	}

	var updateData map[string]interface{}
	if err := json.Unmarshal(drugData, &updateData); err != nil {
		return fmt.Errorf("failed to unmarshal drug data: %v", err)
	}

	_, err = ls.Supabase.Update("drugs", drug.ID, updateData)
	return err
}

// GetDrug retrieves a drug record from the database
func (ls *LedgerStorage) GetDrug(drugID string) (*models.Drug, error) {
	where := map[string]interface{}{"id": drugID}
	result, err := ls.Supabase.Select("drugs", "*", where)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("drug not found: %s", drugID)
	}

	// Convert map to Drug struct
	drugData, err := json.Marshal(result[0])
	if err != nil {
		return nil, err
	}

	var drug models.Drug
	if err := json.Unmarshal(drugData, &drug); err != nil {
		return nil, err
	}

	return &drug, nil
}

// InsertDrugStatusUpdate inserts a drug status update into the database
func (ls *LedgerStorage) InsertDrugStatusUpdate(update *models.DrugStatusUpdate) error {
	_, err := ls.Supabase.Insert("drug_status_updates", update)
	return err
}

// GetDrugStatusUpdates retrieves all status updates for a drug
func (ls *LedgerStorage) GetDrugStatusUpdates(drugID string) ([]models.DrugStatusUpdate, error) {
	where := map[string]interface{}{"drug_id": drugID}
	result, err := ls.Supabase.Select("drug_status_updates", "*", where)
	if err != nil {
		return nil, err
	}

	// Convert maps to DrugStatusUpdate structs
	updates := make([]models.DrugStatusUpdate, len(result))
	for i, item := range result {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &updates[i]); err != nil {
			return nil, err
		}
	}

	return updates, nil
}

// InsertShipment inserts a shipment record into the database
func (ls *LedgerStorage) InsertShipment(shipment *models.Shipment) error {
	_, err := ls.Supabase.Insert("shipments", shipment)
	return err
}

// UpdateShipment updates a shipment record in the database
func (ls *LedgerStorage) UpdateShipment(shipment *models.Shipment) error {
	// Convert shipment to map for update
	shipmentData, err := json.Marshal(shipment)
	if err != nil {
		return fmt.Errorf("failed to marshal shipment: %v", err)
	}

	var updateData map[string]interface{}
	if err := json.Unmarshal(shipmentData, &updateData); err != nil {
		return fmt.Errorf("failed to unmarshal shipment data: %v", err)
	}

	_, err = ls.Supabase.Update("shipments", shipment.ID, updateData)
	return err
}

// GetShipment retrieves a shipment record from the database
func (ls *LedgerStorage) GetShipment(shipmentID string) (*models.Shipment, error) {
	where := map[string]interface{}{"id": shipmentID}
	result, err := ls.Supabase.Select("shipments", "*", where)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("shipment not found: %s", shipmentID)
	}

	// Convert map to Shipment struct
	shipmentData, err := json.Marshal(result[0])
	if err != nil {
		return nil, err
	}

	var shipment models.Shipment
	if err := json.Unmarshal(shipmentData, &shipment); err != nil {
		return nil, err
	}

	return &shipment, nil
}

// InsertShipmentStatusUpdate inserts a shipment status update into the database
func (ls *LedgerStorage) InsertShipmentStatusUpdate(update *models.ShipmentStatusUpdate) error {
	_, err := ls.Supabase.Insert("shipment_status_updates", update)
	return err
}

// GetShipmentStatusUpdates retrieves all status updates for a shipment
func (ls *LedgerStorage) GetShipmentStatusUpdates(shipmentID string) ([]models.ShipmentStatusUpdate, error) {
	where := map[string]interface{}{"shipment_id": shipmentID}
	result, err := ls.Supabase.Select("shipment_status_updates", "*", where)
	if err != nil {
		return nil, err
	}

	// Convert maps to ShipmentStatusUpdate structs
	updates := make([]models.ShipmentStatusUpdate, len(result))
	for i, item := range result {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &updates[i]); err != nil {
			return nil, err
		}
	}

	return updates, nil
}
