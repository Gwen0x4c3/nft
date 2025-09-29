package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// SignatureVerifier handles cryptographic signature operations
type SignatureVerifier struct{}

// NewSignatureVerifier creates a new signature verifier instance
func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{}
}

// VerifyEthereumSignature verifies Ethereum wallet signature
func (sv *SignatureVerifier) VerifyEthereumSignature(walletAddr, signature, message string) error {
	// Validate wallet address format
	if !common.IsHexAddress(walletAddr) {
		return errors.New("invalid wallet address format")
	}

	// Decode signature
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	if len(sigBytes) != 65 {
		return errors.New("signature must be 65 bytes")
	}

	// Prepare message hash (Ethereum signed message format)
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Recover public key from signature
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27 // Transform recovery id
	}

	pubKey, err := crypto.SigToPub(messageHash.Bytes(), sigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Compare addresses (case-insensitive)
	if !strings.EqualFold(recoveredAddr.Hex(), walletAddr) {
		return errors.New("signature does not match wallet address")
	}

	return nil
}

// IsValidEthereumAddress checks if the provided string is a valid Ethereum address
func (sv *SignatureVerifier) IsValidEthereumAddress(address string) bool {
	return common.IsHexAddress(address)
}

// RecoverAddressFromSignature recovers the Ethereum address from a signature
func (sv *SignatureVerifier) RecoverAddressFromSignature(signature, message string) (string, error) {
	// Decode signature
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return "", fmt.Errorf("invalid signature format: %w", err)
	}

	if len(sigBytes) != 65 {
		return "", errors.New("signature must be 65 bytes")
	}

	// Prepare message hash (Ethereum signed message format)
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Recover public key from signature
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27 // Transform recovery id
	}

	pubKey, err := crypto.SigToPub(messageHash.Bytes(), sigBytes)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr.Hex(), nil
}

// HashMessage creates a Keccak256 hash of the message with Ethereum prefix
func (sv *SignatureVerifier) HashMessage(message string) []byte {
	return crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message))).Bytes()
}

// NonceGenerator handles secure nonce generation
type NonceGenerator struct{}

// NewNonceGenerator creates a new nonce generator instance
func NewNonceGenerator() *NonceGenerator {
	return &NonceGenerator{}
}

// Generate generates a cryptographically secure random nonce
func (ng *NonceGenerator) Generate() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateWithLength generates a nonce with specified byte length
func (ng *NonceGenerator) GenerateWithLength(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// MessageGenerator handles standard authentication message generation
type MessageGenerator struct {
	nonceGen *NonceGenerator
}

// NewMessageGenerator creates a new message generator instance
func NewMessageGenerator() *MessageGenerator {
	return &MessageGenerator{
		nonceGen: NewNonceGenerator(),
	}
}

// GenerateAuthMessage generates a standard authentication message
func (mg *MessageGenerator) GenerateAuthMessage(action string) string {
	nonce := mg.nonceGen.Generate()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("Please sign this message to %s: %s (timestamp: %d)", action, nonce, timestamp)
}

// GenerateCustomMessage generates a custom message with nonce and timestamp
func (mg *MessageGenerator) GenerateCustomMessage(template string, params ...interface{}) string {
	nonce := mg.nonceGen.Generate()
	timestamp := time.Now().Unix()
	
	// Add nonce and timestamp to params
	allParams := append(params, nonce, timestamp)
	return fmt.Sprintf(template, allParams...)
}

// Global instances for convenience (optional)
var (
	DefaultSignatureVerifier = NewSignatureVerifier()
	DefaultNonceGenerator    = NewNonceGenerator()
	DefaultMessageGenerator  = NewMessageGenerator()
)

// Package-level convenience functions
func VerifyEthereumSignature(walletAddr, signature, message string) error {
	return DefaultSignatureVerifier.VerifyEthereumSignature(walletAddr, signature, message)
}

func IsValidEthereumAddress(address string) bool {
	return DefaultSignatureVerifier.IsValidEthereumAddress(address)
}

func RecoverAddressFromSignature(signature, message string) (string, error) {
	return DefaultSignatureVerifier.RecoverAddressFromSignature(signature, message)
}

func GenerateNonce() string {
	return DefaultNonceGenerator.Generate()
}

func GenerateAuthMessage(action string) string {
	return DefaultMessageGenerator.GenerateAuthMessage(action)
}