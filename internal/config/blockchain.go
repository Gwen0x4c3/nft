package config

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BlockchainConfig holds the blockchain configuration
type BlockchainConfig struct {
	RPCEndpoint          string
	WSEndpoint           string
	ChainID              int64
	PrivateKey           string
	ContractAddresses    map[string]string
	GasLimit             uint64
	GasPrice             *big.Int
	MaxGasPrice          *big.Int
	ConfirmationBlocks   uint64
	TransactionTimeout   time.Duration
	RetryAttempts        int
	RetryDelay           time.Duration
}

// LoadBlockchainConfig loads blockchain configuration from environment variables
func LoadBlockchainConfig() *BlockchainConfig {
	gasPrice := big.NewInt(getEnvInt64("BLOCKCHAIN_GAS_PRICE", 20000000000)) // 20 Gwei
	maxGasPrice := big.NewInt(getEnvInt64("BLOCKCHAIN_MAX_GAS_PRICE", 100000000000)) // 100 Gwei

	return &BlockchainConfig{
		RPCEndpoint:        getEnvString("BLOCKCHAIN_RPC_ENDPOINT", "https://sepolia.infura.io/v3/your-project-id"),
		WSEndpoint:         getEnvString("BLOCKCHAIN_WS_ENDPOINT", "wss://sepolia.infura.io/ws/v3/your-project-id"),
		ChainID:            getEnvInt64("BLOCKCHAIN_CHAIN_ID", 11155111), // Sepolia testnet
		PrivateKey:         getEnvString("BLOCKCHAIN_PRIVATE_KEY", ""),
		ContractAddresses:  loadContractAddresses(),
		GasLimit:           uint64(getEnvInt64("BLOCKCHAIN_GAS_LIMIT", 300000)),
		GasPrice:           gasPrice,
		MaxGasPrice:        maxGasPrice,
		ConfirmationBlocks: uint64(getEnvInt64("BLOCKCHAIN_CONFIRMATION_BLOCKS", 3)),
		TransactionTimeout: time.Duration(getEnvInt64("BLOCKCHAIN_TX_TIMEOUT_MINUTES", 5)) * time.Minute,
		RetryAttempts:      getEnvInt("BLOCKCHAIN_RETRY_ATTEMPTS", 3),
		RetryDelay:         time.Duration(getEnvInt64("BLOCKCHAIN_RETRY_DELAY_SECONDS", 10)) * time.Second,
	}
}

// Contract names
const (
	ContractNFT     = "nft"
	ContractAuction = "auction"
	ContractMarket  = "marketplace"
)

// loadContractAddresses loads contract addresses from environment variables
func loadContractAddresses() map[string]string {
	return map[string]string{
		ContractNFT:     getEnvString("CONTRACT_NFT_ADDRESS", ""),
		ContractAuction: getEnvString("CONTRACT_AUCTION_ADDRESS", ""),
		ContractMarket:  getEnvString("CONTRACT_MARKET_ADDRESS", ""),
	}
}

// BlockchainClient wraps the Ethereum client with additional functionality
type BlockchainClient struct {
	client     *ethclient.Client
	wsClient   *ethclient.Client
	config     *BlockchainConfig
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
}

// NewBlockchainClient creates a new blockchain client
func NewBlockchainClient(config *BlockchainConfig) (*BlockchainClient, error) {
	// Connect to RPC endpoint
	client, err := ethclient.Dial(config.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum RPC: %w", err)
	}

	// Connect to WebSocket endpoint for event subscriptions
	wsClient, err := ethclient.Dial(config.WSEndpoint)
	if err != nil {
		log.Printf("Warning: failed to connect to WebSocket endpoint: %v", err)
		// WebSocket is optional, continue without it
		wsClient = nil
	}

	bc := &BlockchainClient{
		client:   client,
		wsClient: wsClient,
		config:   config,
	}

	// Load private key if provided
	if config.PrivateKey != "" {
		if err := bc.loadPrivateKey(config.PrivateKey); err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Int64() != config.ChainID {
		return nil, fmt.Errorf("chain ID mismatch: expected %d, got %d", config.ChainID, chainID.Int64())
	}

	log.Printf("Blockchain client connected: Chain ID %d", chainID.Int64())
	return bc, nil
}

// loadPrivateKey loads the private key from hex string
func (bc *BlockchainClient) loadPrivateKey(privateKeyHex string) error {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKey)

	bc.privateKey = privateKey
	bc.publicKey = publicKey
	bc.address = address

	log.Printf("Loaded wallet address: %s", address.Hex())
	return nil
}

// GetAddress returns the wallet address
func (bc *BlockchainClient) GetAddress() common.Address {
	return bc.address
}

// GetBalance returns the ETH balance of the wallet
func (bc *BlockchainClient) GetBalance(ctx context.Context) (*big.Int, error) {
	return bc.client.BalanceAt(ctx, bc.address, nil)
}

// GetNonce returns the current nonce for the wallet
func (bc *BlockchainClient) GetNonce(ctx context.Context) (uint64, error) {
	return bc.client.PendingNonceAt(ctx, bc.address)
}

// EstimateGas estimates gas for a transaction
func (bc *BlockchainClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return bc.client.EstimateGas(ctx, msg)
}

// SuggestGasPrice returns the suggested gas price
func (bc *BlockchainClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := bc.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// Cap gas price at maximum
	if gasPrice.Cmp(bc.config.MaxGasPrice) > 0 {
		return bc.config.MaxGasPrice, nil
	}

	return gasPrice, nil
}

// CreateTransactor creates a transactor for contract interaction
func (bc *BlockchainClient) CreateTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	if bc.privateKey == nil {
		return nil, fmt.Errorf("private key not loaded")
	}

	nonce, err := bc.GetNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := bc.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	chainID := big.NewInt(bc.config.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(bc.privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = bc.config.GasLimit
	auth.GasPrice = gasPrice
	auth.Context = ctx

	return auth, nil
}

// SendTransaction sends a transaction and waits for confirmation
func (bc *BlockchainClient) SendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	if err := bc.client.SendTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("Transaction sent: %s", tx.Hash().Hex())

	// Wait for confirmation
	receipt, err := bc.WaitForReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	if receipt.Status == 0 {
		return receipt, fmt.Errorf("transaction reverted: %s", tx.Hash().Hex())
	}

	log.Printf("Transaction confirmed: %s (Block: %d, Gas Used: %d)", 
		tx.Hash().Hex(), receipt.BlockNumber.Uint64(), receipt.GasUsed)

	return receipt, nil
}

// WaitForReceipt waits for transaction receipt
func (bc *BlockchainClient) WaitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, bc.config.TransactionTimeout)
	defer cancel()

	for {
		receipt, err := bc.client.TransactionReceipt(ctx, txHash)
		if err == nil {
			// Wait for confirmations
			currentBlock, err := bc.client.BlockNumber(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get current block: %w", err)
			}

			confirmations := currentBlock - receipt.BlockNumber.Uint64()
			if confirmations >= bc.config.ConfirmationBlocks {
				return receipt, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue polling
		}
	}
}

// GetContractAddress returns the address for a named contract
func (bc *BlockchainClient) GetContractAddress(contractName string) (common.Address, error) {
	address, exists := bc.config.ContractAddresses[contractName]
	if !exists || address == "" {
		return common.Address{}, fmt.Errorf("contract address not configured: %s", contractName)
	}

	if !common.IsHexAddress(address) {
		return common.Address{}, fmt.Errorf("invalid contract address: %s", address)
	}

	return common.HexToAddress(address), nil
}

// SubscribeToLogs subscribes to contract logs (requires WebSocket connection)
func (bc *BlockchainClient) SubscribeToLogs(ctx context.Context, query ethereum.FilterQuery, logs chan types.Log) (ethereum.Subscription, error) {
	if bc.wsClient == nil {
		return nil, fmt.Errorf("WebSocket client not available")
	}

	return bc.wsClient.SubscribeFilterLogs(ctx, query, logs)
}

// GetLatestBlock returns the latest block number
func (bc *BlockchainClient) GetLatestBlock(ctx context.Context) (uint64, error) {
	return bc.client.BlockNumber(ctx)
}

// GetBlockByNumber returns a block by number
func (bc *BlockchainClient) GetBlockByNumber(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	return bc.client.BlockByNumber(ctx, blockNumber)
}

// HealthCheck checks blockchain connectivity
func (bc *BlockchainClient) HealthCheck(ctx context.Context) error {
	_, err := bc.client.ChainID(ctx)
	return err
}

// Close closes the blockchain client connections
func (bc *BlockchainClient) Close() {
	if bc.client != nil {
		bc.client.Close()
	}
	if bc.wsClient != nil {
		bc.wsClient.Close()
	}
}

// RetryWithBackoff executes a function with exponential backoff
func (bc *BlockchainClient) RetryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error
	delay := bc.config.RetryDelay

	for attempt := 0; attempt < bc.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay *= 2 // Exponential backoff
			}
		}

		if err := operation(); err != nil {
			lastErr = err
			log.Printf("Blockchain operation failed (attempt %d/%d): %v", 
				attempt+1, bc.config.RetryAttempts, err)
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", bc.config.RetryAttempts, lastErr)
}

// Helper functions

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// CreateTestBlockchainClient creates a blockchain client for testing
func CreateTestBlockchainClient() (*BlockchainClient, error) {
	testConfig := &BlockchainConfig{
		RPCEndpoint:        getEnvString("TEST_BLOCKCHAIN_RPC", "http://localhost:8545"),
		WSEndpoint:         getEnvString("TEST_BLOCKCHAIN_WS", "ws://localhost:8545"),
		ChainID:            getEnvInt64("TEST_CHAIN_ID", 1337), // Local testnet
		PrivateKey:         getEnvString("TEST_PRIVATE_KEY", ""),
		ContractAddresses:  make(map[string]string),
		GasLimit:           300000,
		GasPrice:           big.NewInt(20000000000),
		MaxGasPrice:        big.NewInt(100000000000),
		ConfirmationBlocks: 1,
		TransactionTimeout: 30 * time.Second,
		RetryAttempts:      2,
		RetryDelay:         2 * time.Second,
	}

	return NewBlockchainClient(testConfig)
}

// Event structures for common blockchain events
type NFTMintedEvent struct {
	TokenID   *big.Int       `json:"token_id"`
	To        common.Address `json:"to"`
	TokenURI  string         `json:"token_uri"`
	BlockNum  uint64         `json:"block_number"`
	TxHash    common.Hash    `json:"tx_hash"`
	Timestamp time.Time      `json:"timestamp"`
}

type NFTTransferEvent struct {
	TokenID   *big.Int       `json:"token_id"`
	From      common.Address `json:"from"`
	To        common.Address `json:"to"`
	BlockNum  uint64         `json:"block_number"`
	TxHash    common.Hash    `json:"tx_hash"`
	Timestamp time.Time      `json:"timestamp"`
}

type AuctionCreatedEvent struct {
	AuctionID    *big.Int       `json:"auction_id"`
	TokenID      *big.Int       `json:"token_id"`
	Seller       common.Address `json:"seller"`
	StartPrice   *big.Int       `json:"start_price"`
	ReservePrice *big.Int       `json:"reserve_price"`
	EndTime      *big.Int       `json:"end_time"`
	BlockNum     uint64         `json:"block_number"`
	TxHash       common.Hash    `json:"tx_hash"`
	Timestamp    time.Time      `json:"timestamp"`
}

type BidPlacedEvent struct {
	AuctionID *big.Int       `json:"auction_id"`
	Bidder    common.Address `json:"bidder"`
	Amount    *big.Int       `json:"amount"`
	BlockNum  uint64         `json:"block_number"`
	TxHash    common.Hash    `json:"tx_hash"`
	Timestamp time.Time      `json:"timestamp"`
}