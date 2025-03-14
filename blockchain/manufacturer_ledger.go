package blockchain

import (
	"encoding/json"
	"fmt"
	"time"
)

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

// Status represents a status update for a drug
type Status struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Details   string `json:"details,omitempty"`
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

// NewManufacturerLedger creates a new manufacturer ledger
func NewManufacturerLedger(manufacturerID string) *ManufacturerLedger {
	return &ManufacturerLedger{
		ManufacturerID: manufacturerID,
		Drugs:          []DrugRecord{},
		Shipments:      []ShipmentRecord{},
		LastUpdated:    time.Now().Format(time.RFC3339),
	}
}

// AddDrugRecord adds a new drug record to the manufacturer's ledger
func (ml *ManufacturerLedger) AddDrugRecord(drugID string, status string) error {
	// Check if drug already exists
	for _, drug := range ml.Drugs {
		if drug.DrugID == drugID {
			return fmt.Errorf("drug %s already exists in ledger", drugID)
		}
	}

	// Create new drug record
	drugRecord := DrugRecord{
		DrugID:        drugID,
		Status:        status,
		CreatedAt:     time.Now().Format(time.RFC3339),
		CurrentStatus: status,
		History: []Status{
			{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	ml.Drugs = append(ml.Drugs, drugRecord)
	ml.LastUpdated = time.Now().Format(time.RFC3339)
	return nil
}

// UpdateDrugStatus updates a drug's status in the manufacturer's ledger
func (ml *ManufacturerLedger) UpdateDrugStatus(drugID string, newStatus string, details string) error {
	for i := range ml.Drugs {
		if ml.Drugs[i].DrugID == drugID {
			ml.Drugs[i].CurrentStatus = newStatus
			ml.Drugs[i].History = append(ml.Drugs[i].History, Status{
				Status:    newStatus,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			ml.LastUpdated = time.Now().Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("drug %s not found in ledger", drugID)
}

// AddShipmentRecord adds a new shipment record to the manufacturer's ledger
func (ml *ManufacturerLedger) AddShipmentRecord(shipmentID string, drugID string, status string) error {
	// Check if shipment already exists
	for _, shipment := range ml.Shipments {
		if shipment.ShipmentID == shipmentID {
			return fmt.Errorf("shipment %s already exists in ledger", shipmentID)
		}
	}

	// Create new shipment record
	shipmentRecord := ShipmentRecord{
		ShipmentID:    shipmentID,
		DrugID:        drugID,
		Status:        status,
		CreatedAt:     time.Now().Format(time.RFC3339),
		CurrentStatus: status,
		History: []Status{
			{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	ml.Shipments = append(ml.Shipments, shipmentRecord)
	ml.LastUpdated = time.Now().Format(time.RFC3339)
	return nil
}

// UpdateShipmentStatus updates a shipment's status in the manufacturer's ledger
func (ml *ManufacturerLedger) UpdateShipmentStatus(shipmentID string, newStatus string, details string) error {
	for i := range ml.Shipments {
		if ml.Shipments[i].ShipmentID == shipmentID {
			ml.Shipments[i].CurrentStatus = newStatus
			ml.Shipments[i].History = append(ml.Shipments[i].History, Status{
				Status:    newStatus,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			ml.LastUpdated = time.Now().Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("shipment %s not found in ledger", shipmentID)
}

// GetDrugHistory retrieves the complete history of a drug
func (ml *ManufacturerLedger) GetDrugHistory(drugID string) ([]Status, error) {
	for _, drug := range ml.Drugs {
		if drug.DrugID == drugID {
			return drug.History, nil
		}
	}
	return nil, fmt.Errorf("drug %s not found in ledger", drugID)
}

// GetShipmentHistory retrieves the complete history of a shipment
func (ml *ManufacturerLedger) GetShipmentHistory(shipmentID string) ([]Status, error) {
	for _, shipment := range ml.Shipments {
		if shipment.ShipmentID == shipmentID {
			return shipment.History, nil
		}
	}
	return nil, fmt.Errorf("shipment %s not found in ledger", shipmentID)
}

// ToJSON converts the manufacturer ledger to JSON
func (ml *ManufacturerLedger) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ml, "", "  ")
}

// FromJSON creates a manufacturer ledger from JSON
func FromJSON(data []byte) (*ManufacturerLedger, error) {
	var ml ManufacturerLedger
	if err := json.Unmarshal(data, &ml); err != nil {
		return nil, err
	}
	return &ml, nil
}
