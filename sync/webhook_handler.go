package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// WebhookPayload represents the payload received from Supabase webhooks
type WebhookPayload struct {
	Type      string                 `json:"type"`
	Table     string                 `json:"table"`
	Record    map[string]interface{} `json:"record"`
	OldRecord map[string]interface{} `json:"old_record,omitempty"`
	Schema    string                 `json:"schema"`
	Timestamp time.Time              `json:"timestamp"`
}

// WebhookHandler handles incoming webhooks from Supabase
type WebhookHandler struct {
	SyncService        *SyncService
	Secret             string // For webhook verification
	MaxRetries         int    // Maximum number of retries for failed processing
	TransactionTracker *TransactionTracker
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(syncService *SyncService, secret string) (*WebhookHandler, error) {
	// Initialize transaction tracker
	tracker, err := NewTransactionTracker(syncService.Storage.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize transaction tracker: %v", err)
	}

	return &WebhookHandler{
		SyncService:        syncService,
		Secret:             secret,
		MaxRetries:         3, // Default to 3 retries
		TransactionTracker: tracker,
	}, nil
}

// RegisterRoutes registers webhook routes with the Fiber app
func (wh *WebhookHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/api/webhooks/supabase", wh.HandleWebhook)
}

// HandleWebhook processes incoming webhook events from Supabase
func (wh *WebhookHandler) HandleWebhook(c *fiber.Ctx) error {
	// Verify webhook signature if secret is set
	if wh.Secret != "" {
		signature := c.Get("X-Webhook-Signature")
		log.Printf("Received signature: %s", signature)
		log.Printf("Secret expected: %s", wh.Secret)

		if !wh.verifySignature(signature, c.Body()) {
			log.Printf("Invalid webhook signature received from %s", c.IP())
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid signature",
			})
		}
		log.Printf("Signature verification successful")
	}

	// Parse webhook payload
	var payload WebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payload",
		})
	}

	// Log webhook receipt
	log.Printf("Received webhook event: type=%s, table=%s, timestamp=%v",
		payload.Type, payload.Table, payload.Timestamp)

	// Process the webhook based on the event type
	switch payload.Type {
	case "INSERT", "UPDATE":
		go wh.processRecordWithRetry(payload.Table, payload.Record)
	case "DELETE":
		go wh.processDeleteWithRetry(payload.Table, payload.OldRecord)
	default:
		log.Printf("Unhandled webhook event type: %s", payload.Type)
	}

	// Acknowledge receipt of the webhook
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Webhook received and processing started",
	})
}

// processRecordWithRetry processes a record with retry logic
func (wh *WebhookHandler) processRecordWithRetry(tableName string, record map[string]interface{}) {
	var err error
	for i := 0; i < wh.MaxRetries; i++ {
		err = wh.processRecord(tableName, record)
		if err == nil {
			return
		}
		log.Printf("Attempt %d/%d failed to process record: %v", i+1, wh.MaxRetries, err)
		time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
	}
	log.Printf("Failed to process record after %d attempts: %v", wh.MaxRetries, err)
}

// processDeleteWithRetry processes a delete event with retry logic
func (wh *WebhookHandler) processDeleteWithRetry(tableName string, record map[string]interface{}) {
	var err error
	for i := 0; i < wh.MaxRetries; i++ {
		err = wh.processDelete(tableName, record)
		if err == nil {
			return
		}
		log.Printf("Attempt %d/%d failed to process delete: %v", i+1, wh.MaxRetries, err)
		time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
	}
	log.Printf("Failed to process delete after %d attempts: %v", wh.MaxRetries, err)
}

// processRecord processes a record from a webhook event
func (wh *WebhookHandler) processRecord(tableName string, record map[string]interface{}) error {
	log.Printf("Processing %s record from webhook", tableName)

	// Determine record ID and create transaction based on table type
	var recordID string
	var txHash string

	switch tableName {
	case "drugs":
		if id, ok := record["drug_id"].(string); ok {
			recordID = id

			// Check if record already has a blockchain transaction ID
			if txID, ok := record["blockchain_tx_id"].(string); ok && txID != "" && txID != "pending" {
				txHash = txID
			} else {
				// Create blockchain transaction for drug
				txHash = wh.SyncService.createDrugTransaction(record)
			}
		}
	case "shipments":
		if id, ok := record["shipment_id"].(string); ok {
			recordID = id

			// Check if record already has a blockchain transaction ID
			if txID, ok := record["blockchain_tx_id"].(string); ok && txID != "" && txID != "pending" {
				txHash = txID
			} else {
				// Create blockchain transaction for shipment
				txHash = wh.SyncService.createShipmentTransaction(record)
			}
		}
	default:
		return fmt.Errorf("unhandled table in webhook: %s", tableName)
	}

	if recordID == "" {
		return fmt.Errorf("no valid record ID found for table %s", tableName)
	}

	// Check if transaction has already been processed
	if txHash != "" && wh.TransactionTracker.IsTransactionProcessed(txHash) {
		log.Printf("Transaction %s has already been processed, skipping record", txHash)
		return nil
	}

	// Save record to data_records directory
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s_%s.json", tableName, recordID, timestamp)
	filePath := fmt.Sprintf("%s/%s", wh.SyncService.Storage.DataDir, filename)

	// Add blockchain transaction ID to record if available
	if txHash != "" {
		record["blockchain_tx_id"] = txHash
	}

	// Marshal record to JSON
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal record: %v", err)
	}

	// Write to file
	if err := wh.SyncService.Storage.WriteFile(filePath, data); err != nil {
		return fmt.Errorf("failed to write record to file: %v", err)
	}

	// Mark transaction as processed if we have a hash
	if txHash != "" {
		if err := wh.TransactionTracker.MarkTransactionProcessed(txHash); err != nil {
			log.Printf("Warning: Failed to mark transaction %s as processed: %v", txHash, err)
		}
	}

	log.Printf("Successfully saved record to %s", filePath)

	// Update last sync time for this table
	wh.SyncService.SyncLock.Lock()
	wh.SyncService.SyncStatusMap[tableName] = time.Now()
	wh.SyncService.SyncLock.Unlock()

	return nil
}

// processDelete handles DELETE events from Supabase
func (wh *WebhookHandler) processDelete(tableName string, record map[string]interface{}) error {
	log.Printf("Processing DELETE event for %s", tableName)

	var recordID string
	switch tableName {
	case "drugs":
		if id, ok := record["drug_id"].(string); ok {
			recordID = id
		}
	case "shipments":
		if id, ok := record["shipment_id"].(string); ok {
			recordID = id
		}
	default:
		return fmt.Errorf("unhandled table in delete webhook: %s", tableName)
	}

	if recordID == "" {
		return fmt.Errorf("no valid record ID found for table %s", tableName)
	}

	// Create a deletion record
	deletionRecord := map[string]interface{}{
		"deleted_at":      time.Now().Format(time.RFC3339),
		"original_record": record,
	}

	// Save deletion record
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s_%s_deleted.json", tableName, recordID, timestamp)
	filePath := fmt.Sprintf("%s/%s", wh.SyncService.Storage.DataDir, filename)

	data, err := json.MarshalIndent(deletionRecord, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deletion record: %v", err)
	}

	if err := wh.SyncService.Storage.WriteFile(filePath, data); err != nil {
		return fmt.Errorf("failed to write deletion record: %v", err)
	}

	log.Printf("Successfully saved deletion record to %s", filePath)

	// Update last sync time for this table
	wh.SyncService.SyncLock.Lock()
	wh.SyncService.SyncStatusMap[tableName] = time.Now()
	wh.SyncService.SyncLock.Unlock()

	return nil
}

// verifySignature verifies the webhook signature
func (wh *WebhookHandler) verifySignature(signature string, payload []byte) bool {
	if signature == "" {
		log.Printf("Warning: Empty signature received")
		return false
	}

	// Simple string comparison of the signature with the webhook secret
	return signature == wh.Secret
}
