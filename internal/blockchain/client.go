package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"nft-platform/internal/config"
	"nft-platform/internal/service"
)

// Client provides blockchain interaction capabilities
type Client struct {
	ethClient     *ethclient.Client
	wsClient      *ethclient.Client
	config        *config.BlockchainConfig
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	fromAddress   common.Address
	chainID       *big.Int
	contractAddrs map[string]common.Address
	contractABIs  map[string]*abi.ABI
	mu            sync.RWMutex
	connected     bool
	eventSubs     map[string]ethereum.Subscription
	subMu         sync.RWMutex
}

// NewClient creates a new blockchain client
func NewClient(cfg *config.BlockchainConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("blockchain config is required")
	}

	// Connect to Ethereum RPC endpoint
	ethClient, err := ethclient.Dial(cfg.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum RPC: %w", err)
	}

	// Connect to WebSocket endpoint for event subscriptions
	var wsClient *ethclient.Client
	if cfg.WSEndpoint != "" {
		wsClient, err = ethclient.Dial(cfg.WSEndpoint)
		if err != nil {
			// Log warning but continue - WebSocket is optional
			fmt.Printf("Warning: failed to connect to WebSocket endpoint: %v\n", err)
		}
	}

	// Parse private key if provided
	var privateKey *ecdsa.PrivateKey
	var publicKey *ecdsa.PublicKey
	var fromAddress common.Address

	if cfg.PrivateKey != "" {
		// Remove '0x' prefix if present
		privateKeyHex := strings.TrimPrefix(cfg.PrivateKey, "0x")
		privateKey, err = crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}

		publicKey = privateKey.Public().(*ecdsa.PublicKey)
		fromAddress = crypto.PubkeyToAddress(*publicKey)
	}

	// Verify chain ID
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Int64() != cfg.ChainID {
		return nil, fmt.Errorf("chain ID mismatch: expected %d, got %d", cfg.ChainID, chainID.Int64())
	}

	// Initialize contract addresses
	contractAddrs := make(map[string]common.Address)
	for name, addr := range cfg.ContractAddresses {
		if addr != "" && common.IsHexAddress(addr) {
			contractAddrs[name] = common.HexToAddress(addr)
		}
	}

	client := &Client{
		ethClient:     ethClient,
		wsClient:      wsClient,
		config:        cfg,
		privateKey:    privateKey,
		publicKey:     publicKey,
		fromAddress:   fromAddress,
		chainID:       chainID,
		contractAddrs: contractAddrs,
		contractABIs:  make(map[string]*abi.ABI),
		connected:     true,
		eventSubs:     make(map[string]ethereum.Subscription),
	}

	// Load contract ABIs
	if err := client.loadContractABIs(); err != nil {
		return nil, fmt.Errorf("failed to load contract ABIs: %w", err)
	}

	return client, nil
}

// loadContractABIs loads the contract ABIs for interaction
func (c *Client) loadContractABIs() error {
	// ERC-721 NFT contract ABI (simplified version with essential methods)
	nftABI, err := abi.JSON(strings.NewReader(nftContractABI))
	if err != nil {
		return fmt.Errorf("failed to parse NFT ABI: %w", err)
	}
	c.contractABIs[service.ContractTypeNFT] = &nftABI

	// Auction contract ABI (simplified version)
	auctionABI, err := abi.JSON(strings.NewReader(auctionContractABI))
	if err != nil {
		return fmt.Errorf("failed to parse Auction ABI: %w", err)
	}
	c.contractABIs[service.ContractTypeAuction] = &auctionABI

	return nil
}

// MintNFT mints a new NFT on the blockchain
func (c *Client) MintNFT(ctx context.Context, recipient string, tokenURI string) (*service.BlockchainTransaction, error) {
	if !c.connected {
		return nil, fmt.Errorf("blockchain client not connected")
	}

	if c.privateKey == nil {
		return nil, fmt.Errorf("private key not configured")
	}

	contractAddr, ok := c.contractAddrs[service.ContractTypeNFT]
	if !ok {
		return nil, fmt.Errorf("NFT contract address not configured")
	}

	// Validate recipient address
	if !common.IsHexAddress(recipient) {
		return nil, fmt.Errorf("invalid recipient address: %s", recipient)
	}
	recipientAddr := common.HexToAddress(recipient)

	// Get contract ABI
	contractABI, ok := c.contractABIs[service.ContractTypeNFT]
	if !ok {
		return nil, fmt.Errorf("NFT contract ABI not loaded")
	}

	// Pack the mint function call data
	data, err := contractABI.Pack("mint", recipientAddr, tokenURI)
	if err != nil {
		return nil, fmt.Errorf("failed to pack mint data: %w", err)
	}

	// Send transaction
	tx, err := c.sendTransaction(ctx, &contractAddr, big.NewInt(0), data, c.config.GasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to send mint transaction: %w", err)
	}

	// Wait for confirmation
	receipt, err := c.waitForReceipt(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mint transaction receipt: %w", err)
	}

	return c.receiptToTransaction(receipt), nil
}

// TransferNFT transfers an NFT from one address to another
func (c *Client) TransferNFT(ctx context.Context, tokenID string, from, to string) (*service.BlockchainTransaction, error) {
	if !c.connected {
		return nil, fmt.Errorf("blockchain client not connected")
	}

	if c.privateKey == nil {
		return nil, fmt.Errorf("private key not configured")
	}

	contractAddr, ok := c.contractAddrs[service.ContractTypeNFT]
	if !ok {
		return nil, fmt.Errorf("NFT contract address not configured")
	}

	// Validate addresses
	if !common.IsHexAddress(from) {
		return nil, fmt.Errorf("invalid from address: %s", from)
	}
	if !common.IsHexAddress(to) {
		return nil, fmt.Errorf("invalid to address: %s", to)
	}

	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	// Parse token ID
	tokenIDBigInt := new(big.Int)
	if _, ok := tokenIDBigInt.SetString(tokenID, 10); !ok {
		return nil, fmt.Errorf("invalid token ID: %s", tokenID)
	}

	// Get contract ABI
	contractABI, ok := c.contractABIs[service.ContractTypeNFT]
	if !ok {
		return nil, fmt.Errorf("NFT contract ABI not loaded")
	}

	// Pack the transferFrom function call data
	data, err := contractABI.Pack("transferFrom", fromAddr, toAddr, tokenIDBigInt)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transfer data: %w", err)
	}

	// Send transaction
	tx, err := c.sendTransaction(ctx, &contractAddr, big.NewInt(0), data, c.config.GasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to send transfer transaction: %w", err)
	}

	// Wait for confirmation
	receipt, err := c.waitForReceipt(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer transaction receipt: %w", err)
	}

	return c.receiptToTransaction(receipt), nil
}

// GetTokenOwner retrieves the owner of a specific token
func (c *Client) GetTokenOwner(ctx context.Context, tokenID string) (string, error) {
	if !c.connected {
		return "", fmt.Errorf("blockchain client not connected")
	}

	contractAddr, ok := c.contractAddrs[service.ContractTypeNFT]
	if !ok {
		return "", fmt.Errorf("NFT contract address not configured")
	}

	// Parse token ID
	tokenIDBigInt := new(big.Int)
	if _, ok := tokenIDBigInt.SetString(tokenID, 10); !ok {
		return "", fmt.Errorf("invalid token ID: %s", tokenID)
	}

	// Get contract ABI
	contractABI, ok := c.contractABIs[service.ContractTypeNFT]
	if !ok {
		return "", fmt.Errorf("NFT contract ABI not loaded")
	}

	// Pack the ownerOf function call data
	data, err := contractABI.Pack("ownerOf", tokenIDBigInt)
	if err != nil {
		return "", fmt.Errorf("failed to pack ownerOf data: %w", err)
	}

	// Call contract
	result, err := c.ethClient.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call ownerOf: %w", err)
	}

	// Unpack result
	var owner common.Address
	err = contractABI.UnpackIntoInterface(&owner, "ownerOf", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack ownerOf result: %w", err)
	}

	return owner.Hex(), nil
}

// GetNextTokenID retrieves the next available token ID
func (c *Client) GetNextTokenID(ctx context.Context) (string, error) {
	if !c.connected {
		return "", fmt.Errorf("blockchain client not connected")
	}

	contractAddr, ok := c.contractAddrs[service.ContractTypeNFT]
	if !ok {
		return "", fmt.Errorf("NFT contract address not configured")
	}

	// Get contract ABI
	contractABI, ok := c.contractABIs[service.ContractTypeNFT]
	if !ok {
		return "", fmt.Errorf("NFT contract ABI not loaded")
	}

	// Pack the totalSupply function call data
	data, err := contractABI.Pack("totalSupply")
	if err != nil {
		return "", fmt.Errorf("failed to pack totalSupply data: %w", err)
	}

	// Call contract
	result, err := c.ethClient.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call totalSupply: %w", err)
	}

	// Unpack result
	var totalSupply *big.Int
	err = contractABI.UnpackIntoInterface(&totalSupply, "totalSupply", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack totalSupply result: %w", err)
	}

	// Next token ID is totalSupply + 1 (assuming 0-indexed or sequential)
	nextTokenID := new(big.Int).Add(totalSupply, big.NewInt(1))
	return nextTokenID.String(), nil
}

// PlaceBid places a bid on an auction
func (c *Client) PlaceBid(ctx context.Context, tokenID string, bidder string, amount interface{}) (*service.BlockchainTransaction, error) {
	if !c.connected {
		return nil, fmt.Errorf("blockchain client not connected")
	}

	if c.privateKey == nil {
		return nil, fmt.Errorf("private key not configured")
	}

	contractAddr, ok := c.contractAddrs[service.ContractTypeAuction]
	if !ok {
		return nil, fmt.Errorf("Auction contract address not configured")
	}

	// Validate bidder address
	if !common.IsHexAddress(bidder) {
		return nil, fmt.Errorf("invalid bidder address: %s", bidder)
	}

	// Parse token ID
	tokenIDBigInt := new(big.Int)
	if _, ok := tokenIDBigInt.SetString(tokenID, 10); !ok {
		return nil, fmt.Errorf("invalid token ID: %s", tokenID)
	}

	// Convert amount to big.Int
	var bidAmount *big.Int
	switch v := amount.(type) {
	case *big.Int:
		bidAmount = v
	case string:
		bidAmount = new(big.Int)
		if _, ok := bidAmount.SetString(v, 10); !ok {
			return nil, fmt.Errorf("invalid bid amount: %s", v)
		}
	case int64:
		bidAmount = big.NewInt(v)
	case float64:
		bidAmount = big.NewInt(int64(v))
	default:
		return nil, fmt.Errorf("unsupported bid amount type: %T", amount)
	}

	// Get contract ABI
	contractABI, ok := c.contractABIs[service.ContractTypeAuction]
	if !ok {
		return nil, fmt.Errorf("Auction contract ABI not loaded")
	}

	// Pack the bid function call data
	data, err := contractABI.Pack("placeBid", tokenIDBigInt)
	if err != nil {
		return nil, fmt.Errorf("failed to pack placeBid data: %w", err)
	}

	// Send transaction with bid amount as value
	tx, err := c.sendTransaction(ctx, &contractAddr, bidAmount, data, c.config.GasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to send bid transaction: %w", err)
	}

	// Wait for confirmation
	receipt, err := c.waitForReceipt(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid transaction receipt: %w", err)
	}

	return c.receiptToTransaction(receipt), nil
}

// SetContractAddress sets the address of a contract
func (c *Client) SetContractAddress(contractType string, address string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !common.IsHexAddress(address) {
		return fmt.Errorf("invalid contract address: %s", address)
	}

	c.contractAddrs[contractType] = common.HexToAddress(address)
	return nil
}

// GetContractAddress retrieves the address of a contract
func (c *Client) GetContractAddress(contractType string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if addr, ok := c.contractAddrs[contractType]; ok {
		return addr.Hex()
	}
	return ""
}

// GetTransaction retrieves a transaction by hash
func (c *Client) GetTransaction(ctx context.Context, txHash string) (*service.BlockchainTransaction, error) {
	if !c.connected {
		return nil, fmt.Errorf("blockchain client not connected")
	}

	hash := common.HexToHash(txHash)

	// Get transaction
	tx, isPending, err := c.ethClient.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if isPending {
		// Transaction is still pending
		return &service.BlockchainTransaction{
			Hash:   txHash,
			Status: false,
		}, nil
	}

	// Get transaction receipt
	receipt, err := c.ethClient.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Get current block number for confirmations
	currentBlock, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	result := c.receiptToTransaction(receipt)
	result.Confirmations = currentBlock - receipt.BlockNumber.Uint64()

	// Add transaction details
	result.Nonce = tx.Nonce()
	result.Data = hex.EncodeToString(tx.Data())
	result.Value = tx.Value().String()
	result.GasLimit = tx.Gas()

	return result, nil
}

// WaitForConfirmation waits for a transaction to reach the specified number of confirmations
func (c *Client) WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) error {
	if !c.connected {
		return fmt.Errorf("blockchain client not connected")
	}

	hash := common.HexToHash(txHash)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(c.config.TransactionTimeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("transaction confirmation timeout")
		case <-ticker.C:
			receipt, err := c.ethClient.TransactionReceipt(ctx, hash)
			if err != nil {
				continue
			}

			currentBlock, err := c.ethClient.BlockNumber(ctx)
			if err != nil {
				continue
			}

			confirmedBlocks := currentBlock - receipt.BlockNumber.Uint64()
			if confirmedBlocks >= confirmations {
				if receipt.Status == 0 {
					return fmt.Errorf("transaction failed")
				}
				return nil
			}
		}
	}
}

// EstimateGas estimates the gas required for an operation
func (c *Client) EstimateGas(ctx context.Context, operation string, params map[string]interface{}) (uint64, error) {
	if !c.connected {
		return 0, fmt.Errorf("blockchain client not connected")
	}

	// Provide estimates based on operation type
	switch operation {
	case "mint":
		return 200000, nil
	case "transfer":
		return 100000, nil
	case "bid":
		return 150000, nil
	case "approve":
		return 50000, nil
	default:
		return 100000, nil
	}
}

// GetBalance retrieves the balance of an address
func (c *Client) GetBalance(ctx context.Context, address string) (string, error) {
	if !c.connected {
		return "", fmt.Errorf("blockchain client not connected")
	}

	if !common.IsHexAddress(address) {
		return "", fmt.Errorf("invalid address: %s", address)
	}

	addr := common.HexToAddress(address)
	balance, err := c.ethClient.BalanceAt(ctx, addr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	return balance.String(), nil
}

// GetNonce retrieves the nonce for an address
func (c *Client) GetNonce(ctx context.Context, address string) (uint64, error) {
	if !c.connected {
		return 0, fmt.Errorf("blockchain client not connected")
	}

	if !common.IsHexAddress(address) {
		return 0, fmt.Errorf("invalid address: %s", address)
	}

	addr := common.HexToAddress(address)
	nonce, err := c.ethClient.PendingNonceAt(ctx, addr)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	return nonce, nil
}

// SubscribeToEvents subscribes to blockchain events
func (c *Client) SubscribeToEvents(ctx context.Context, eventTypes []string) (<-chan service.BlockchainEvent, error) {
	if !c.connected {
		return nil, fmt.Errorf("blockchain client not connected")
	}

	if c.wsClient == nil {
		return nil, fmt.Errorf("WebSocket client not available for event subscriptions")
	}

	eventChan := make(chan service.BlockchainEvent, 100)

	// Create filter query for all contract addresses
	addresses := make([]common.Address, 0)
	for _, addr := range c.contractAddrs {
		addresses = append(addresses, addr)
	}

	// Create event type topics
	topics := make([][]common.Hash, 0)
	if len(eventTypes) > 0 {
		eventTopics := make([]common.Hash, 0)
		for _, eventType := range eventTypes {
			// Convert event type to topic hash (simplified)
			eventTopics = append(eventTopics, crypto.Keccak256Hash([]byte(eventType+"(address,address,uint256)")))
		}
		topics = append(topics, eventTopics)
	}

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    topics,
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := c.wsClient.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Store subscription
	c.subMu.Lock()
	subscriptionKey := fmt.Sprintf("events_%d", time.Now().UnixNano())
	c.eventSubs[subscriptionKey] = sub
	c.subMu.Unlock()

	// Process logs in a goroutine
	go func() {
		defer close(eventChan)
		defer func() {
			c.subMu.Lock()
			delete(c.eventSubs, subscriptionKey)
			c.subMu.Unlock()
		}()

		for {
			select {
			case err := <-sub.Err():
				fmt.Printf("Event subscription error: %v\n", err)
				return
			case vLog := <-logs:
				event := c.logToEvent(vLog)
				select {
				case eventChan <- event:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, nil
}

// GetEvents retrieves historical events from the blockchain
func (c *Client) GetEvents(ctx context.Context, fromBlock, toBlock uint64, eventTypes []string) ([]service.BlockchainEvent, error) {
	if !c.connected {
		return nil, fmt.Errorf("blockchain client not connected")
	}

	// Create filter query
	addresses := make([]common.Address, 0)
	for _, addr := range c.contractAddrs {
		addresses = append(addresses, addr)
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: addresses,
	}

	// Get logs
	logs, err := c.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	// Convert logs to events
	events := make([]service.BlockchainEvent, 0, len(logs))
	for _, vLog := range logs {
		events = append(events, c.logToEvent(vLog))
	}

	return events, nil
}

// IsConnected checks if the client is connected to the blockchain
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetChainID retrieves the chain ID
func (c *Client) GetChainID(ctx context.Context) (int64, error) {
	if !c.connected {
		return 0, fmt.Errorf("blockchain client not connected")
	}

	return c.chainID.Int64(), nil
}

// GetBlockNumber retrieves the current block number
func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	if !c.connected {
		return 0, fmt.Errorf("blockchain client not connected")
	}

	blockNumber, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}

	return blockNumber, nil
}

// Close closes the blockchain client connections
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false

	// Close all event subscriptions
	c.subMu.Lock()
	for _, sub := range c.eventSubs {
		sub.Unsubscribe()
	}
	c.eventSubs = make(map[string]ethereum.Subscription)
	c.subMu.Unlock()

	// Close clients
	if c.ethClient != nil {
		c.ethClient.Close()
	}
	if c.wsClient != nil {
		c.wsClient.Close()
	}
}

// sendTransaction sends a transaction to the blockchain
func (c *Client) sendTransaction(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasLimit uint64) (*types.Transaction, error) {
	nonce, err := c.ethClient.PendingNonceAt(ctx, c.fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get current gas price
	gasPrice, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Cap gas price at max
	if c.config.MaxGasPrice != nil && gasPrice.Cmp(c.config.MaxGasPrice) > 0 {
		gasPrice = c.config.MaxGasPrice
	}

	// Use configured gas price if set
	if c.config.GasPrice != nil && c.config.GasPrice.Cmp(big.NewInt(0)) > 0 {
		gasPrice = c.config.GasPrice
	}

	// Create transaction
	tx := types.NewTransaction(nonce, *to, value, gasLimit, gasPrice, data)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = c.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}

// waitForReceipt waits for a transaction receipt
func (c *Client) waitForReceipt(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(c.config.TransactionTimeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("transaction receipt timeout")
		case <-ticker.C:
			receipt, err := c.ethClient.TransactionReceipt(ctx, tx.Hash())
			if err == nil {
				return receipt, nil
			}
		}
	}
}

// receiptToTransaction converts a transaction receipt to BlockchainTransaction
func (c *Client) receiptToTransaction(receipt *types.Receipt) *service.BlockchainTransaction {
	logs := make([]service.BlockchainLog, 0, len(receipt.Logs))
	for _, log := range receipt.Logs {
		topics := make([]string, 0, len(log.Topics))
		for _, topic := range log.Topics {
			topics = append(topics, topic.Hex())
		}

		logs = append(logs, service.BlockchainLog{
			Address:     log.Address.Hex(),
			Topics:      topics,
			Data:        hex.EncodeToString(log.Data),
			BlockNumber: log.BlockNumber,
			TxHash:      log.TxHash.Hex(),
			TxIndex:     log.TxIndex,
			BlockHash:   log.BlockHash.Hex(),
			Index:       log.Index,
			Removed:     log.Removed,
		})
	}

	var toAddr string
	if receipt.ContractAddress != (common.Address{}) {
		toAddr = receipt.ContractAddress.Hex()
	}

	return &service.BlockchainTransaction{
		Hash:             receipt.TxHash.Hex(),
		Status:           receipt.Status == 1,
		BlockNumber:      receipt.BlockNumber.Uint64(),
		BlockHash:        receipt.BlockHash.Hex(),
		TransactionIndex: receipt.TransactionIndex,
		To:               toAddr,
		GasUsed:          receipt.GasUsed,
		Logs:             logs,
		Confirmations:    0, // Will be calculated by caller if needed
	}
}

// logToEvent converts a blockchain log to BlockchainEvent
func (c *Client) logToEvent(vLog types.Log) service.BlockchainEvent {
	eventType := "Unknown"
	var tokenID, from, to string

	// Parse common event types based on topics
	if len(vLog.Topics) > 0 {
		eventSig := vLog.Topics[0].Hex()

		// Transfer event signature
		transferSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex()
		// Approval event signature
		approvalSig := crypto.Keccak256Hash([]byte("Approval(address,address,uint256)")).Hex()

		switch eventSig {
		case transferSig:
			eventType = service.EventTypeTransfer
			if len(vLog.Topics) >= 4 {
				from = common.HexToAddress(vLog.Topics[1].Hex()).Hex()
				to = common.HexToAddress(vLog.Topics[2].Hex()).Hex()
				tokenID = vLog.Topics[3].Big().String()
			}
		case approvalSig:
			eventType = service.EventTypeApproval
			if len(vLog.Topics) >= 4 {
				from = common.HexToAddress(vLog.Topics[1].Hex()).Hex()
				to = common.HexToAddress(vLog.Topics[2].Hex()).Hex()
				tokenID = vLog.Topics[3].Big().String()
			}
		}
	}

	data := make(map[string]interface{})
	data["raw_data"] = hex.EncodeToString(vLog.Data)

	return service.BlockchainEvent{
		Type:        eventType,
		Contract:    vLog.Address.Hex(),
		TokenID:     tokenID,
		From:        from,
		To:          to,
		Data:        data,
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		LogIndex:    vLog.Index,
		Timestamp:   time.Now().Unix(), // Would need to fetch block for actual timestamp
	}
}

// Contract ABIs (simplified versions with essential methods)
const nftContractABI = `[
	{
		"inputs": [
			{"name": "to", "type": "address"},
			{"name": "tokenURI", "type": "string"}
		],
		"name": "mint",
		"outputs": [{"name": "tokenId", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "from", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "tokenId", "type": "uint256"}
		],
		"name": "transferFrom",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "tokenId", "type": "uint256"}],
		"name": "ownerOf",
		"outputs": [{"name": "owner", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalSupply",
		"outputs": [{"name": "supply", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": true, "name": "tokenId", "type": "uint256"}
		],
		"name": "Transfer",
		"type": "event"
	}
]`

const auctionContractABI = `[
	{
		"inputs": [{"name": "tokenId", "type": "uint256"}],
		"name": "placeBid",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "tokenId", "type": "uint256"},
			{"name": "startingPrice", "type": "uint256"},
			{"name": "duration", "type": "uint256"}
		],
		"name": "createAuction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "tokenId", "type": "uint256"}],
		"name": "finalizeAuction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "tokenId", "type": "uint256"},
			{"indexed": true, "name": "bidder", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "BidPlaced",
		"type": "event"
	}
]`
