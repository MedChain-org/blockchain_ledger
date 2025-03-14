package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ankit/blockchain_ledger/models"
	"github.com/ankit/blockchain_ledger/sync"
	"github.com/google/uuid"
)

// Handler represents the HTTP handler for the API
type Handler struct {
	ledgerManager models.LedgerManager
	syncService   *sync.SyncService
}

// NewHandler creates a new handler
func NewHandler(ledgerManager models.LedgerManager, syncService *sync.SyncService) *Handler {
	return &Handler{
		ledgerManager: ledgerManager,
		syncService:   syncService,
	}
}

// SetupRoutes sets up the HTTP routes for the API
func SetupRoutes(ledgerManager models.LedgerManager, syncService *sync.SyncService) {
	handler := NewHandler(ledgerManager, syncService)

	// Drug routes
	http.HandleFunc("/api/drugs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateDrug(w, r)
		case http.MethodGet:
			handler.GetDrugs(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/drugs/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetDrug(w, r)
		case http.MethodPut:
			handler.RevertDrug(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Shipment routes
	http.HandleFunc("/api/shipments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateShipment(w, r)
		case http.MethodGet:
			handler.GetShipments(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/shipments/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetShipment(w, r)
		case http.MethodPut:
			handler.UpdateShipmentStatus(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Verification routes
	http.HandleFunc("/api/verify/", handler.VerifyDrug)

	// Sync routes
	http.HandleFunc("/api/sync/status", handler.GetSyncStatus)
	http.HandleFunc("/api/sync/force", handler.ForceSync)

	// Webhook route for real-time sync
	http.HandleFunc("/api/webhook", handler.HandleWebhook)
}

// CreateDrug handles the creation of a new drug
func (h *Handler) CreateDrug(w http.ResponseWriter, r *http.Request) {
	var params models.CreateDrugParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate drug ID if not provided
	if params.DrugID == "" {
		params.DrugID = uuid.New().String()
	}

	verificationHash, err := h.ledgerManager.CreateDrug(&params)
	if err != nil {
		log.Printf("Error creating drug: %v", err)
		http.Error(w, "Failed to create drug", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"drug_id":           params.DrugID,
		"verification_hash": verificationHash,
		"message":           "Drug created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetDrugs handles the retrieval of all drugs
func (h *Handler) GetDrugs(w http.ResponseWriter, r *http.Request) {
	// This would typically query the database for all drugs
	// For now, we'll return a placeholder response
	response := map[string]interface{}{
		"message": "Get drugs endpoint",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDrug handles the retrieval of a specific drug
func (h *Handler) GetDrug(w http.ResponseWriter, r *http.Request) {
	// Extract drug ID from URL
	drugID := r.URL.Path[len("/api/drugs/"):]
	if drugID == "" {
		http.Error(w, "Drug ID is required", http.StatusBadRequest)
		return
	}

	// This would typically query the database for the specific drug
	// For now, we'll return a placeholder response
	response := map[string]interface{}{
		"drug_id": drugID,
		"message": "Get drug endpoint",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevertDrug handles the reversion of a drug
func (h *Handler) RevertDrug(w http.ResponseWriter, r *http.Request) {
	// Extract drug ID from URL
	drugID := r.URL.Path[len("/api/drugs/"):]
	if drugID == "" {
		http.Error(w, "Drug ID is required", http.StatusBadRequest)
		return
	}

	var params models.RevertDrugParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params.DrugID = drugID

	if err := h.ledgerManager.RevertDrug(&params); err != nil {
		log.Printf("Error reverting drug: %v", err)
		http.Error(w, "Failed to revert drug", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"drug_id": drugID,
		"message": "Drug reverted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateShipment handles the creation of a new shipment
func (h *Handler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	var params models.CreateShipmentParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate shipment ID if not provided
	if params.ShipmentID == "" {
		params.ShipmentID = uuid.New().String()
	}

	txHash, err := h.ledgerManager.CreateShipment(&params)
	if err != nil {
		log.Printf("Error creating shipment: %v", err)
		http.Error(w, "Failed to create shipment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"shipment_id":      params.ShipmentID,
		"blockchain_tx_id": txHash,
		"message":          "Shipment created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetShipments handles the retrieval of all shipments
func (h *Handler) GetShipments(w http.ResponseWriter, r *http.Request) {
	// This would typically query the database for all shipments
	// For now, we'll return a placeholder response
	response := map[string]interface{}{
		"message": "Get shipments endpoint",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetShipment handles the retrieval of a specific shipment
func (h *Handler) GetShipment(w http.ResponseWriter, r *http.Request) {
	// Extract shipment ID from URL
	shipmentID := r.URL.Path[len("/api/shipments/"):]
	if shipmentID == "" {
		http.Error(w, "Shipment ID is required", http.StatusBadRequest)
		return
	}

	// This would typically query the database for the specific shipment
	// For now, we'll return a placeholder response
	response := map[string]interface{}{
		"shipment_id": shipmentID,
		"message":     "Get shipment endpoint",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateShipmentStatus handles the update of a shipment's status
func (h *Handler) UpdateShipmentStatus(w http.ResponseWriter, r *http.Request) {
	// Extract shipment ID from URL
	shipmentID := r.URL.Path[len("/api/shipments/"):]
	if shipmentID == "" {
		http.Error(w, "Shipment ID is required", http.StatusBadRequest)
		return
	}

	var params models.UpdateShipmentStatusParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params.ShipmentID = shipmentID

	if err := h.ledgerManager.UpdateShipmentStatus(&params); err != nil {
		log.Printf("Error updating shipment status: %v", err)
		http.Error(w, "Failed to update shipment status", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"shipment_id": shipmentID,
		"status":      params.Status,
		"message":     "Shipment status updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyDrug handles the verification of a drug
func (h *Handler) VerifyDrug(w http.ResponseWriter, r *http.Request) {
	// Extract drug ID from URL
	drugID := r.URL.Path[len("/api/verify/"):]
	if drugID == "" {
		http.Error(w, "Drug ID is required", http.StatusBadRequest)
		return
	}

	isVerified, err := h.ledgerManager.VerifyDrug(drugID)
	if err != nil {
		log.Printf("Error verifying drug: %v", err)
		http.Error(w, "Failed to verify drug", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"drug_id":     drugID,
		"is_verified": isVerified,
		"message":     "Drug verification completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSyncStatus handles the retrieval of sync status
func (h *Handler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	status := h.syncService.GetSyncStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ForceSync handles the forcing of a sync operation
func (h *Handler) ForceSync(w http.ResponseWriter, r *http.Request) {
	if err := h.syncService.ForceSync(); err != nil {
		log.Printf("Error forcing sync: %v", err)
		http.Error(w, "Failed to force sync", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Sync operation started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SupabaseWebhookPayload represents the structure of a Supabase webhook payload
type SupabaseWebhookPayload struct {
	Type      string                 `json:"type"`
	Table     string                 `json:"table"`
	Schema    string                 `json:"schema"`
	Record    map[string]interface{} `json:"record"`
	OldRecord map[string]interface{} `json:"old_record,omitempty"`
}

// HandleWebhook handles webhook requests for real-time sync
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request details
	log.Printf("Received webhook request:")
	log.Printf("  Method: %s", r.Method)
	log.Printf("  URL: %s", r.URL.String())
	log.Printf("  RemoteAddr: %s", r.RemoteAddr)
	log.Printf("  Headers: %v", r.Header)

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log the raw payload for debugging
	log.Printf("Webhook payload: %s", string(body))

	// Parse the webhook payload
	var payload SupabaseWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	// Log the parsed payload
	log.Printf("Webhook type: %s, Table: %s, Schema: %s", payload.Type, payload.Table, payload.Schema)
	log.Printf("Record: %v", payload.Record)
	if payload.OldRecord != nil {
		log.Printf("Old Record: %v", payload.OldRecord)
	}

	// Handle different types of events
	switch payload.Type {
	case "INSERT":
		log.Printf("New record inserted in table %s", payload.Table)
		// Trigger sync for the affected table
		if err := h.syncService.ForceSync(); err != nil {
			log.Printf("Error syncing after webhook: %v", err)
			http.Error(w, "Failed to sync after webhook", http.StatusInternalServerError)
			return
		}
	case "UPDATE":
		log.Printf("Record updated in table %s", payload.Table)
		// Trigger sync for the affected table
		if err := h.syncService.ForceSync(); err != nil {
			log.Printf("Error syncing after webhook: %v", err)
			http.Error(w, "Failed to sync after webhook", http.StatusInternalServerError)
			return
		}
	case "DELETE":
		log.Printf("Record deleted from table %s", payload.Table)
		// Trigger sync for the affected table
		if err := h.syncService.ForceSync(); err != nil {
			log.Printf("Error syncing after webhook: %v", err)
			http.Error(w, "Failed to sync after webhook", http.StatusInternalServerError)
			return
		}
	default:
		log.Printf("Unknown webhook type: %s", payload.Type)
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Webhook processed successfully",
		"type":    payload.Type,
		"table":   payload.Table,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
