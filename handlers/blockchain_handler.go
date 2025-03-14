package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ankit/blockchain_ledger/storage"
	"github.com/ankit/blockchain_ledger/blockchain"
	"time"
)

// BlockchainHandler handles blockchain-related API endpoints
type BlockchainHandler struct {
	Storage *storage.DataStorage
}

// NewBlockchainHandler creates a new blockchain handler
func NewBlockchainHandler(storage *storage.DataStorage) *BlockchainHandler {
	return &BlockchainHandler{Storage: storage}
}

// GetStatus returns the current status of the blockchain
func (h *BlockchainHandler) GetStatus(c *fiber.Ctx) error {
	// Get blockchain ledger
	ledger, err := h.Storage.GetBlockchainLedger()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve blockchain ledger",
		})
	}

	// Return blockchain status
	return c.JSON(fiber.Map{
		"status": "active",
		"block_height": ledger.BlockHeight,
		"last_updated": ledger.LastUpdated,
		"block_count": len(ledger.Blocks),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// VerifyTransaction verifies a blockchain transaction
func (h *BlockchainHandler) VerifyTransaction(c *fiber.Ctx) error {
	txHash := c.Params("tx_hash")
	if txHash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing transaction hash",
		})
	}

	// Get blockchain ledger
	ledger, err := h.Storage.GetBlockchainLedger()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve blockchain ledger",
		})
	}

	// Find transaction in blockchain
	var txData map[string]interface{}
	var blockHeight int
	var timestamp string

	for _, block := range ledger.Blocks {
		if block.TxHash == txHash {
			txData, _ = block.TxData.(map[string]interface{})
			blockHeight = block.BlockHeight
			timestamp = block.Timestamp
			break
		}
	}

	if txData == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	// Verify transaction
	isValid := blockchain.ValidateTransaction(txHash, txData)

	return c.JSON(fiber.Map{
		"tx_hash": txHash,
		"is_valid": isValid,
		"block_height": blockHeight,
		"timestamp": timestamp,
		"tx_data": txData,
	})
}

// RunConsistencyCheck runs a consistency check on the blockchain
func (h *BlockchainHandler) RunConsistencyCheck(c *fiber.Ctx) error {
	// Run consistency check
	result := h.Storage.RunConsistencyCheck()

	return c.JSON(result)
}