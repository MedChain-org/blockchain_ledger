package models

import "time"

// LedgerStorage defines the interface for storage operations
type LedgerStorage interface {
	// Manufacturer ledger operations
	GetManufacturerLedger(manufacturerID string) (*ManufacturerLedger, error)
	SaveManufacturerLedger(ledger *ManufacturerLedger) error

	// Common ledger operations
	GetCommonLedger() (*CommonLedger, error)
	SaveCommonLedger(ledger *CommonLedger) error

	// Database operations
	InsertDrug(drug *Drug) error
	UpdateDrug(drug *Drug) error
	GetDrug(drugID string) (*Drug, error)

	InsertDrugStatusUpdate(update *DrugStatusUpdate) error
	GetDrugStatusUpdates(drugID string) ([]DrugStatusUpdate, error)

	InsertShipment(shipment *Shipment) error
	UpdateShipment(shipment *Shipment) error
	GetShipment(shipmentID string) (*Shipment, error)

	InsertShipmentStatusUpdate(update *ShipmentStatusUpdate) error
	GetShipmentStatusUpdates(shipmentID string) ([]ShipmentStatusUpdate, error)
}

// BlockchainService defines the interface for blockchain operations
type BlockchainService interface {
	// Transaction operations
	CreateTransaction(txType string, data map[string]interface{}) (string, error)
	GetTransaction(txID string) (*BlockchainTransaction, error)
	VerifyTransaction(txID string) (bool, error)

	// Ledger operations
	RecordDrugCreation(drugID, manufacturerID string, timestamp time.Time) (string, error)
	RecordDrugStatusUpdate(drugID, status, updatedBy string, timestamp time.Time) (string, error)
	RecordShipmentCreation(shipmentID, drugID, manufacturerID, distributorID string, timestamp time.Time) (string, error)
	RecordShipmentStatusUpdate(shipmentID, status, updatedBy string, timestamp time.Time) (string, error)
}

// LedgerManager defines the interface for ledger management operations
type LedgerManager interface {
	// Drug operations
	CreateDrug(params *CreateDrugParams) (string, error)
	RevertDrug(params *RevertDrugParams) error

	// Shipment operations
	CreateShipment(params *CreateShipmentParams) (string, error)
	UpdateShipmentStatus(params *UpdateShipmentStatusParams) error

	// Verification operations
	VerifyDrug(drugID string) (bool, error)
	GetDrugHistory(drugID string) ([]DrugStatusUpdate, error)
	GetShipmentHistory(shipmentID string) ([]ShipmentStatusUpdate, error)
}
