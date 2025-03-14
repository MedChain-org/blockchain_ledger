package models

import "time"

// Drug represents a drug record in the database
type Drug struct {
	ID               string    `json:"id"`
	ManufacturerID   string    `json:"manufacturer_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Status           string    `json:"status"`
	VerificationHash string    `json:"verification_hash"`
	BlockchainTxID   string    `json:"blockchain_tx_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DrugStatusUpdate represents a drug status update in the database
type DrugStatusUpdate struct {
	ID             string    `json:"id"`
	DrugID         string    `json:"drug_id"`
	Status         string    `json:"status"`
	Location       string    `json:"location"`
	UpdatedBy      string    `json:"updated_by"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// Shipment represents a shipment record in the database
type Shipment struct {
	ID             string    `json:"id"`
	DrugID         string    `json:"drug_id"`
	ManufacturerID string    `json:"manufacturer_id"`
	DistributorID  string    `json:"distributor_id"`
	Status         string    `json:"status"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ShipmentStatusUpdate represents a shipment status update in the database
type ShipmentStatusUpdate struct {
	ID             string    `json:"id"`
	ShipmentID     string    `json:"shipment_id"`
	Status         string    `json:"status"`
	Location       string    `json:"location"`
	UpdatedBy      string    `json:"updated_by"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// Prescription represents a prescription record in the database
type Prescription struct {
	ID             string    `json:"id"`
	DrugID         string    `json:"drug_id"`
	PatientID      string    `json:"patient_id"`
	DoctorID       string    `json:"doctor_id"`
	Dosage         string    `json:"dosage"`
	Instructions   string    `json:"instructions"`
	Status         string    `json:"status"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// User represents a user record in the database
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"` // manufacturer, distributor, doctor, patient
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HealthCheck represents a health check record in the database
type HealthCheck struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// BlockchainTransaction represents a blockchain transaction
type BlockchainTransaction struct {
	TransactionID string                 `json:"transaction_id"`
	Type          string                 `json:"type"` // drug_create, drug_update, shipment_create, shipment_update
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
}

// CreateDrugParams represents the parameters for creating a drug
type CreateDrugParams struct {
	DrugID         string `json:"drug_id"`
	ManufacturerID string `json:"manufacturer_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	UserID         string `json:"user_id"`
	Location       string `json:"location"`
}

// CreateShipmentParams represents the parameters for creating a shipment
type CreateShipmentParams struct {
	ShipmentID     string `json:"shipment_id"`
	DrugID         string `json:"drug_id"`
	ManufacturerID string `json:"manufacturer_id"`
	DistributorID  string `json:"distributor_id"`
	UserID         string `json:"user_id"`
	Location       string `json:"location"`
}

// UpdateShipmentStatusParams represents the parameters for updating a shipment status
type UpdateShipmentStatusParams struct {
	ShipmentID string `json:"shipment_id"`
	Status     string `json:"status"`
	UserID     string `json:"user_id"`
	Location   string `json:"location"`
}

// RevertDrugParams represents the parameters for reverting a drug
type RevertDrugParams struct {
	DrugID         string `json:"drug_id"`
	ManufacturerID string `json:"manufacturer_id"`
	Reason         string `json:"reason"`
	UserID         string `json:"user_id"`
	Location       string `json:"location"`
}
