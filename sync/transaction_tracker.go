package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TransactionTracker manages transaction hashes to prevent duplicates
type TransactionTracker struct {
	mu           sync.RWMutex
	TxHashFile   string
	ProcessedTxs map[string]bool
}

// NewTransactionTracker creates a new transaction tracker
func NewTransactionTracker(dataDir string) (*TransactionTracker, error) {
	txHashFile := filepath.Join(dataDir, "processed_transactions.json")

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	tracker := &TransactionTracker{
		TxHashFile:   txHashFile,
		ProcessedTxs: make(map[string]bool),
	}

	// Load existing transaction hashes if file exists
	if err := tracker.loadProcessedTxs(); err != nil {
		return nil, fmt.Errorf("failed to load processed transactions: %v", err)
	}

	return tracker, nil
}

// loadProcessedTxs loads previously processed transaction hashes from file
func (t *TransactionTracker) loadProcessedTxs() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(t.TxHashFile); os.IsNotExist(err) {
		return nil // File doesn't exist yet, that's okay
	}

	// Read file
	data, err := os.ReadFile(t.TxHashFile)
	if err != nil {
		return fmt.Errorf("failed to read transaction hash file: %v", err)
	}

	// Parse JSON
	var txHashes []string
	if err := json.Unmarshal(data, &txHashes); err != nil {
		return fmt.Errorf("failed to parse transaction hashes: %v", err)
	}

	// Add to map
	for _, hash := range txHashes {
		t.ProcessedTxs[hash] = true
	}

	return nil
}

// saveProcessedTxs saves processed transaction hashes to file
func (t *TransactionTracker) saveProcessedTxs() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Convert map keys to slice
	txHashes := make([]string, 0, len(t.ProcessedTxs))
	for hash := range t.ProcessedTxs {
		txHashes = append(txHashes, hash)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(txHashes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transaction hashes: %v", err)
	}

	// Write to file
	if err := os.WriteFile(t.TxHashFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write transaction hashes: %v", err)
	}

	return nil
}

// IsTransactionProcessed checks if a transaction hash has been processed before
func (t *TransactionTracker) IsTransactionProcessed(txHash string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ProcessedTxs[txHash]
}

// MarkTransactionProcessed marks a transaction hash as processed
func (t *TransactionTracker) MarkTransactionProcessed(txHash string) error {
	t.mu.Lock()
	t.ProcessedTxs[txHash] = true
	t.mu.Unlock()

	// Save to file
	return t.saveProcessedTxs()
}
