package models

import (
	"encoding/json"
	"time"
)

// Status represents a status update for a drug or shipment
type Status struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Details   string `json:"details,omitempty"`
}

// ManufacturerLedger represents a manufacturer-specific ledger
type ManufacturerLedger struct {
	ManufacturerID string           `json:"manufacturer_id"`
	Drugs          []DrugRecord     `json:"drugs"`
	Shipments      []ShipmentRecord `json:"shipments"`
	LastUpdated    string           `json:"last_updated"`
}

// DrugRecord represents a drug's status and history
type DrugRecord struct {
	DrugID        string   `json:"drug_id"`
	Status        string   `json:"status"` // created, reverted, in_transit, delivered
	CreatedAt     string   `json:"created_at"`
	RevertedAt    string   `json:"reverted_at,omitempty"`
	CurrentStatus string   `json:"current_status"`
	History       []Status `json:"history"`
}

// ShipmentRecord represents a shipment's status and history
type ShipmentRecord struct {
	ShipmentID    string   `json:"shipment_id"`
	DrugID        string   `json:"drug_id"`
	Status        string   `json:"status"` // created, in_transit, delivered, failed
	CreatedAt     string   `json:"created_at"`
	CurrentStatus string   `json:"current_status"`
	History       []Status `json:"history"`
}

// CommonLedger represents the shared ledger for distributors and users
type CommonLedger struct {
	Drugs       []CommonDrugRecord     `json:"drugs"`
	Shipments   []CommonShipmentRecord `json:"shipments"`
	LastUpdated string                 `json:"last_updated"`
}

// CommonDrugRecord represents a drug's public information in the common ledger
type CommonDrugRecord struct {
	DrugID           string   `json:"drug_id"`
	ManufacturerID   string   `json:"manufacturer_id"`
	Status           string   `json:"status"` // active, reverted, in_transit, delivered
	CreatedAt        string   `json:"created_at"`
	CurrentStatus    string   `json:"current_status"`
	History          []Status `json:"history"`
	VerificationHash string   `json:"verification_hash"`
}

// CommonShipmentRecord represents a shipment's public information in the common ledger
type CommonShipmentRecord struct {
	ShipmentID     string   `json:"shipment_id"`
	DrugID         string   `json:"drug_id"`
	ManufacturerID string   `json:"manufacturer_id"`
	DistributorID  string   `json:"distributor_id"`
	Status         string   `json:"status"` // created, in_transit, delivered, failed
	CreatedAt      string   `json:"created_at"`
	CurrentStatus  string   `json:"current_status"`
	History        []Status `json:"history"`
}

// NewManufacturerLedger creates a new manufacturer ledger
func NewManufacturerLedger(manufacturerID string) *ManufacturerLedger {
	return &ManufacturerLedger{
		ManufacturerID: manufacturerID,
		Drugs:          []DrugRecord{},
		Shipments:      []ShipmentRecord{},
		LastUpdated:    time.Now().Format(time.RFC3339),
	}
}

// NewCommonLedger creates a new common ledger
func NewCommonLedger() *CommonLedger {
	return &CommonLedger{
		Drugs:       []CommonDrugRecord{},
		Shipments:   []CommonShipmentRecord{},
		LastUpdated: time.Now().Format(time.RFC3339),
	}
}

// ToJSON converts the manufacturer ledger to JSON
func (ml *ManufacturerLedger) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ml, "", "  ")
}

// ToJSON converts the common ledger to JSON
func (cl *CommonLedger) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cl, "", "  ")
}
