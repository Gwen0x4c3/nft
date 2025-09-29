package service

import (
	"context"
)

// BlockchainService interface for blockchain interactions
type BlockchainService interface {
	// NFT operations
	MintNFT(ctx context.Context, recipient string, tokenURI string) (*BlockchainTransaction, error)
	TransferNFT(ctx context.Context, tokenID string, from, to string) (*BlockchainTransaction, error)
	GetTokenOwner(ctx context.Context, tokenID string) (string, error)
	GetNextTokenID(ctx context.Context) (string, error)
	
	// Contract operations
	SetContractAddress(contractType string, address string) error
	GetContractAddress(contractType string) string
	
	// Transaction operations
	GetTransaction(ctx context.Context, txHash string) (*BlockchainTransaction, error)
	WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) error
	EstimateGas(ctx context.Context, operation string, params map[string]interface{}) (uint64, error)
	
	// Account operations
	GetBalance(ctx context.Context, address string) (string, error) // Returns balance in Wei as string
	GetNonce(ctx context.Context, address string) (uint64, error)
	
	// Event operations
	SubscribeToEvents(ctx context.Context, eventTypes []string) (<-chan BlockchainEvent, error)
	GetEvents(ctx context.Context, fromBlock, toBlock uint64, eventTypes []string) ([]BlockchainEvent, error)
	
	// Health and status
	IsConnected() bool
	GetChainID(ctx context.Context) (int64, error)
	GetBlockNumber(ctx context.Context) (uint64, error)
}

// BlockchainTransaction represents a blockchain transaction result
type BlockchainTransaction struct {
	Hash            string                 `json:"hash"`
	Status          bool                   `json:"status"`
	BlockNumber     uint64                 `json:"block_number"`
	BlockHash       string                 `json:"block_hash"`
	TransactionIndex uint                  `json:"transaction_index"`
	From            string                 `json:"from"`
	To              string                 `json:"to"`
	Value           string                 `json:"value"` // Wei amount as string
	GasLimit        uint64                 `json:"gas_limit"`
	GasUsed         uint64                 `json:"gas_used"`
	GasPrice        string                 `json:"gas_price"` // Wei amount as string
	Nonce           uint64                 `json:"nonce"`
	Data            string                 `json:"data"`
	Logs            []BlockchainLog        `json:"logs"`
	Timestamp       int64                  `json:"timestamp"`
	Confirmations   uint64                 `json:"confirmations"`
	Error           string                 `json:"error,omitempty"`
}

// BlockchainLog represents a transaction log entry
type BlockchainLog struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber uint64   `json:"block_number"`
	TxHash      string   `json:"transaction_hash"`
	TxIndex     uint     `json:"transaction_index"`
	BlockHash   string   `json:"block_hash"`
	Index       uint     `json:"log_index"`
	Removed     bool     `json:"removed"`
}

// BlockchainEvent represents a blockchain event
type BlockchainEvent struct {
	Type        string                 `json:"type"`
	Contract    string                 `json:"contract"`
	TokenID     string                 `json:"token_id,omitempty"`
	From        string                 `json:"from,omitempty"`
	To          string                 `json:"to,omitempty"`
	Value       string                 `json:"value,omitempty"`
	Data        map[string]interface{} `json:"data"`
	TxHash      string                 `json:"tx_hash"`
	BlockNumber uint64                 `json:"block_number"`
	LogIndex    uint                   `json:"log_index"`
	Timestamp   int64                  `json:"timestamp"`
}

// Contract types
const (
	ContractTypeNFT        = "nft"
	ContractTypeAuction    = "auction"
	ContractTypeMarketplace = "marketplace"
)

// Event types
const (
	EventTypeTransfer     = "Transfer"
	EventTypeMint         = "Mint"
	EventTypeApproval     = "Approval"
	EventTypeApprovalForAll = "ApprovalForAll"
	EventTypeAuctionCreated = "AuctionCreated"
	EventTypeAuctionEnded   = "AuctionEnded"
	EventTypeBidPlaced     = "BidPlaced"
)

// BlockchainError represents blockchain-specific errors
type BlockchainError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (e BlockchainError) Error() string {
	return e.Message
}

// Common blockchain error codes
const (
	ErrCodeInsufficientFunds = 4001
	ErrCodeGasLimitExceeded  = 4002
	ErrCodeInvalidSignature  = 4003
	ErrCodeContractReverted  = 4004
	ErrCodeNonceConflict     = 4005
	ErrCodeNetworkCongestion = 4006
	ErrCodeInvalidAddress    = 4007
	ErrCodeTokenNotFound     = 4008
)

// MockBlockchainService provides a mock implementation for testing
type MockBlockchainService struct {
	connected   bool
	chainID     int64
	blockNumber uint64
	events      chan BlockchainEvent
}

// NewMockBlockchainService creates a new mock blockchain service
func NewMockBlockchainService() *MockBlockchainService {
	return &MockBlockchainService{
		connected:   true,
		chainID:     11155111, // Sepolia testnet
		blockNumber: 1000000,
		events:      make(chan BlockchainEvent, 100),
	}
}

// Implementation of BlockchainService interface for mock
func (m *MockBlockchainService) MintNFT(ctx context.Context, recipient string, tokenURI string) (*BlockchainTransaction, error) {
	return &BlockchainTransaction{
		Hash:         generateMockTxHash(),
		Status:       true,
		BlockNumber:  m.blockNumber + 1,
		BlockHash:    generateMockBlockHash(),
		From:         "0x1234567890123456789012345678901234567890",
		To:           "0x1234567890123456789012345678901234567890", // Contract address
		GasUsed:      200000,
		GasPrice:     "20000000000", // 20 Gwei
		Confirmations: 1,
	}, nil
}

func (m *MockBlockchainService) TransferNFT(ctx context.Context, tokenID string, from, to string) (*BlockchainTransaction, error) {
	return &BlockchainTransaction{
		Hash:         generateMockTxHash(),
		Status:       true,
		BlockNumber:  m.blockNumber + 1,
		BlockHash:    generateMockBlockHash(),
		From:         from,
		To:           "0x1234567890123456789012345678901234567890", // Contract address
		GasUsed:      100000,
		GasPrice:     "20000000000", // 20 Gwei
		Confirmations: 1,
	}, nil
}

func (m *MockBlockchainService) GetTokenOwner(ctx context.Context, tokenID string) (string, error) {
	return "0x1234567890123456789012345678901234567890", nil
}

func (m *MockBlockchainService) GetNextTokenID(ctx context.Context) (string, error) {
	return generateMockTokenID(), nil
}

func (m *MockBlockchainService) SetContractAddress(contractType string, address string) error {
	return nil
}

func (m *MockBlockchainService) GetContractAddress(contractType string) string {
	return "0x1234567890123456789012345678901234567890"
}

func (m *MockBlockchainService) GetTransaction(ctx context.Context, txHash string) (*BlockchainTransaction, error) {
	return &BlockchainTransaction{
		Hash:         txHash,
		Status:       true,
		BlockNumber:  m.blockNumber,
		Confirmations: 5,
	}, nil
}

func (m *MockBlockchainService) WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) error {
	return nil
}

func (m *MockBlockchainService) EstimateGas(ctx context.Context, operation string, params map[string]interface{}) (uint64, error) {
	switch operation {
	case "mint":
		return 200000, nil
	case "transfer":
		return 100000, nil
	default:
		return 50000, nil
	}
}

func (m *MockBlockchainService) GetBalance(ctx context.Context, address string) (string, error) {
	return "1000000000000000000", nil // 1 ETH in Wei
}

func (m *MockBlockchainService) GetNonce(ctx context.Context, address string) (uint64, error) {
	return 42, nil
}

func (m *MockBlockchainService) SubscribeToEvents(ctx context.Context, eventTypes []string) (<-chan BlockchainEvent, error) {
	return m.events, nil
}

func (m *MockBlockchainService) GetEvents(ctx context.Context, fromBlock, toBlock uint64, eventTypes []string) ([]BlockchainEvent, error) {
	return []BlockchainEvent{}, nil
}

func (m *MockBlockchainService) IsConnected() bool {
	return m.connected
}

func (m *MockBlockchainService) GetChainID(ctx context.Context) (int64, error) {
	return m.chainID, nil
}

func (m *MockBlockchainService) GetBlockNumber(ctx context.Context) (uint64, error) {
	return m.blockNumber, nil
}

// Helper functions for mock implementation
func generateMockTxHash() string {
	return "0x" + generateRandomHex(64)
}

func generateMockBlockHash() string {
	return "0x" + generateRandomHex(64)
}

func generateMockTokenID() string {
	return generateRandomHex(32)
}

