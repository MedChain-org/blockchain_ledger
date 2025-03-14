package blockchain

import (
	"encoding/json"
	"fmt"
	"time"
)

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

// NewCommonLedger creates a new common ledger
func NewCommonLedger() *CommonLedger {
	return &CommonLedger{
		Drugs:       []CommonDrugRecord{},
		Shipments:   []CommonShipmentRecord{},
		LastUpdated: time.Now().Format(time.RFC3339),
	}
}

// AddDrugRecord adds a new drug record to the common ledger
func (cl *CommonLedger) AddDrugRecord(drugID string, manufacturerID string, status string, verificationHash string) error {
	// Check if drug already exists
	for _, drug := range cl.Drugs {
		if drug.DrugID == drugID {
			return fmt.Errorf("drug %s already exists in common ledger", drugID)
		}
	}

	// Create new drug record
	drugRecord := CommonDrugRecord{
		DrugID:           drugID,
		ManufacturerID:   manufacturerID,
		Status:           status,
		CreatedAt:        time.Now().Format(time.RFC3339),
		CurrentStatus:    status,
		VerificationHash: verificationHash,
		History: []Status{
			{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	cl.Drugs = append(cl.Drugs, drugRecord)
	cl.LastUpdated = time.Now().Format(time.RFC3339)
	return nil
}

// UpdateDrugStatus updates a drug's status in the common ledger
func (cl *CommonLedger) UpdateDrugStatus(drugID string, newStatus string, details string) error {
	for i := range cl.Drugs {
		if cl.Drugs[i].DrugID == drugID {
			cl.Drugs[i].CurrentStatus = newStatus
			cl.Drugs[i].History = append(cl.Drugs[i].History, Status{
				Status:    newStatus,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			cl.LastUpdated = time.Now().Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("drug %s not found in common ledger", drugID)
}

// AddShipmentRecord adds a new shipment record to the common ledger
func (cl *CommonLedger) AddShipmentRecord(shipmentID string, drugID string, manufacturerID string, distributorID string, status string) error {
	// Check if shipment already exists
	for _, shipment := range cl.Shipments {
		if shipment.ShipmentID == shipmentID {
			return fmt.Errorf("shipment %s already exists in common ledger", shipmentID)
		}
	}

	// Create new shipment record
	shipmentRecord := CommonShipmentRecord{
		ShipmentID:     shipmentID,
		DrugID:         drugID,
		ManufacturerID: manufacturerID,
		DistributorID:  distributorID,
		Status:         status,
		CreatedAt:      time.Now().Format(time.RFC3339),
		CurrentStatus:  status,
		History: []Status{
			{
				Status:    status,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	cl.Shipments = append(cl.Shipments, shipmentRecord)
	cl.LastUpdated = time.Now().Format(time.RFC3339)
	return nil
}

// UpdateShipmentStatus updates a shipment's status in the common ledger
func (cl *CommonLedger) UpdateShipmentStatus(shipmentID string, newStatus string, details string) error {
	for i := range cl.Shipments {
		if cl.Shipments[i].ShipmentID == shipmentID {
			cl.Shipments[i].CurrentStatus = newStatus
			cl.Shipments[i].History = append(cl.Shipments[i].History, Status{
				Status:    newStatus,
				Timestamp: time.Now().Format(time.RFC3339),
				Details:   details,
			})
			cl.LastUpdated = time.Now().Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("shipment %s not found in common ledger", shipmentID)
}

// VerifyDrug checks if a drug exists and is authentic
func (cl *CommonLedger) VerifyDrug(drugID string, verificationHash string) (bool, error) {
	for _, drug := range cl.Drugs {
		if drug.DrugID == drugID {
			return drug.VerificationHash == verificationHash, nil
		}
	}
	return false, fmt.Errorf("drug %s not found in common ledger", drugID)
}

// GetDrugHistory retrieves the complete history of a drug from the common ledger
func (cl *CommonLedger) GetDrugHistory(drugID string) ([]Status, error) {
	for _, drug := range cl.Drugs {
		if drug.DrugID == drugID {
			return drug.History, nil
		}
	}
	return nil, fmt.Errorf("drug %s not found in common ledger", drugID)
}

// GetShipmentHistory retrieves the complete history of a shipment from the common ledger
func (cl *CommonLedger) GetShipmentHistory(shipmentID string) ([]Status, error) {
	for _, shipment := range cl.Shipments {
		if shipment.ShipmentID == shipmentID {
			return shipment.History, nil
		}
	}
	return nil, fmt.Errorf("shipment %s not found in common ledger", shipmentID)
}

// GetDistributorShipments retrieves all shipments for a specific distributor
func (cl *CommonLedger) GetDistributorShipments(distributorID string) []CommonShipmentRecord {
	var shipments []CommonShipmentRecord
	for _, shipment := range cl.Shipments {
		if shipment.DistributorID == distributorID {
			shipments = append(shipments, shipment)
		}
	}
	return shipments
}

// ToJSON converts the common ledger to JSON
func (cl *CommonLedger) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cl, "", "  ")
}

// CommonLedgerFromJSON creates a common ledger from JSON
func CommonLedgerFromJSON(data []byte) (*CommonLedger, error) {
	var cl CommonLedger
	if err := json.Unmarshal(data, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}
