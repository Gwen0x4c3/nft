package validation

import (
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the go-playground validator with custom validation rules
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new custom validator instance
func NewCustomValidator() *CustomValidator {
	validate := validator.New()
	
	// Register custom validation rules
	validate.RegisterValidation("eth_addr", validateEthereumAddress)
	validate.RegisterValidation("alphanum", validateAlphanumeric)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("wallet_signature", validateWalletSignature)
	validate.RegisterValidation("hex_string", validateHexString)
	validate.RegisterValidation("nft_status", validateNFTStatus)
	validate.RegisterValidation("auction_status", validateAuctionStatus)
	validate.RegisterValidation("transfer_type", validateTransferType)
	validate.RegisterValidation("notification_type", validateNotificationType)
	
	return &CustomValidator{
		validator: validate,
	}
}

// Struct validates a struct
func (cv *CustomValidator) Struct(s interface{}) error {
	return cv.validator.Struct(s)
}

// Var validates a single variable
func (cv *CustomValidator) Var(field interface{}, tag string) error {
	return cv.validator.Var(field, tag)
}

// Custom validation functions

// validateEthereumAddress validates Ethereum address format
func validateEthereumAddress(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	return common.IsHexAddress(address)
}

// validateAlphanumeric validates alphanumeric characters with underscores
func validateAlphanumeric(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true // Let required validator handle empty strings
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, str)
	return matched
}

// validateUsername validates username format (alphanumeric, dash, underscore)
func validateUsername(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true // Let required validator handle empty strings
	}
	// Username: 3-30 characters, alphanumeric, dash, underscore, no consecutive special chars
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$`, str)
	return matched && len(str) >= 3 && len(str) <= 30
}

// validateWalletSignature validates wallet signature format (hex string, 130-132 chars)
func validateWalletSignature(fl validator.FieldLevel) bool {
	signature := fl.Field().String()
	if signature == "" {
		return true // Let required validator handle empty strings
	}
	
	// Remove 0x prefix if present
	cleanSig := strings.TrimPrefix(signature, "0x")
	
	// Check if it's a valid hex string and proper length (65 bytes = 130 hex chars)
	matched, _ := regexp.MatchString(`^[a-fA-F0-9]{130}$`, cleanSig)
	return matched
}

// validateHexString validates hex string format
func validateHexString(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true // Let required validator handle empty strings
	}
	
	// Remove 0x prefix if present
	cleanStr := strings.TrimPrefix(str, "0x")
	matched, _ := regexp.MatchString(`^[a-fA-F0-9]+$`, cleanStr)
	return matched
}

// validateNFTStatus validates NFT status values
func validateNFTStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"draft", "minted", "listed", "sold", "burned"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// validateAuctionStatus validates auction status values
func validateAuctionStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"active", "ended", "cancelled"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// validateTransferType validates transfer type values
func validateTransferType(fl validator.FieldLevel) bool {
	transferType := fl.Field().String()
	validTypes := []string{"mint", "transfer", "sale", "auction"}
	for _, validType := range validTypes {
		if transferType == validType {
			return true
		}
	}
	return false
}

// validateNotificationType validates notification type values
func validateNotificationType(fl validator.FieldLevel) bool {
	notificationType := fl.Field().String()
	validTypes := []string{"bid_placed", "auction_won", "auction_ended", "nft_sold", "nft_received", "follow", "system"}
	for _, validType := range validTypes {
		if notificationType == validType {
			return true
		}
	}
	return false
}

// Static validation functions for direct use

// IsValidEthereumAddress checks if address is a valid Ethereum address
func IsValidEthereumAddress(address string) bool {
	return common.IsHexAddress(address)
}

// IsValidUsername checks if username follows platform rules
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$`, username)
	return matched
}

// IsValidEmail checks if email format is valid (basic check)
func IsValidEmail(email string) bool {
	// Basic email validation - for production, use more robust validation
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

// IsValidURL checks if URL format is valid (basic check)
func IsValidURL(url string) bool {
	// Basic URL validation
	matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, url)
	return matched
}

// IsValidHexString checks if string is valid hexadecimal
func IsValidHexString(str string) bool {
	cleanStr := strings.TrimPrefix(str, "0x")
	matched, _ := regexp.MatchString(`^[a-fA-F0-9]+$`, cleanStr)
	return matched && len(cleanStr)%2 == 0 // Even length
}

// IsValidWalletSignature checks if signature format is valid
func IsValidWalletSignature(signature string) bool {
	cleanSig := strings.TrimPrefix(signature, "0x")
	matched, _ := regexp.MatchString(`^[a-fA-F0-9]{130}$`, cleanSig)
	return matched
}

// ValidationTags provides common validation tag combinations
type ValidationTags struct{}

// Common validation tag constants
const (
	TagRequired     = "required"
	TagEmail        = "email"
	TagEthAddress   = "eth_addr"
	TagUsername     = "username"
	TagAlphanum     = "alphanum"
	TagSignature    = "wallet_signature"
	TagHex          = "hex_string"
	TagURL          = "url"
	TagNFTStatus    = "nft_status"
	TagAuctionStatus = "auction_status"
	TagTransferType = "transfer_type"
	TagNotifType    = "notification_type"
)

// Tag combinations for common use cases
var (
	TagRequiredEmail       = TagRequired + "," + TagEmail
	TagRequiredEthAddr     = TagRequired + "," + TagEthAddress
	TagRequiredUsername    = TagRequired + "," + TagUsername
	TagRequiredSignature   = TagRequired + "," + TagSignature
	TagOptionalURL         = "omitempty," + TagURL
	TagOptionalUsername    = "omitempty," + TagUsername
)

// Global validator instance for convenience
var DefaultValidator = NewCustomValidator()

// Package-level convenience functions
func Struct(s interface{}) error {
	return DefaultValidator.Struct(s)
}

func Var(field interface{}, tag string) error {
	return DefaultValidator.Var(field, tag)
}