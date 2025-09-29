package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT token claims structure
type Claims struct {
	UserID       uint      `json:"user_id"`
	WalletAddr   string    `json:"wallet_address"`
	Username     string    `json:"username"`
	IsVerified   bool      `json:"is_verified"`
	TokenType    TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// JWTManager handles JWT token operations
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

// NewJWTManager creates a new JWT manager instance
func NewJWTManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer string) *JWTManager {
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
	}
}

// NewJWTManagerWithKeys creates a JWT manager with generated keys (for development)
func NewJWTManagerWithKeys(issuer string) (*JWTManager, error) {
	privateKey, publicKey, err := GenerateRSAKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA keys: %w", err)
	}
	
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
	}, nil
}

// GenerateTokenPair creates access and refresh tokens for a user
func (j *JWTManager) GenerateTokenPair(userID uint, walletAddr, username string, isVerified bool) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(24 * time.Hour)       // 24 hours
	refreshExpiry := now.Add(30 * 24 * time.Hour) // 30 days

	// Create access token
	accessClaims := &Claims{
		UserID:     userID,
		WalletAddr: walletAddr,
		Username:   username,
		IsVerified: isVerified,
		TokenType:  AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token
	refreshClaims := &Claims{
		UserID:     userID,
		WalletAddr: walletAddr,
		Username:   username,
		IsVerified: isVerified,
		TokenType:  RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(accessExpiry.Sub(now).Seconds()),
	}, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.TokenType != expectedType {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExportPrivateKeyPEM exports RSA private key as PEM string
func (j *JWTManager) ExportPrivateKeyPEM() string {
	privateKeyDER := x509.MarshalPKCS1PrivateKey(j.privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	}
	return string(pem.EncodeToMemory(privateKeyBlock))
}

// ExportPublicKeyPEM exports RSA public key as PEM string
func (j *JWTManager) ExportPublicKeyPEM() string {
	publicKeyDER, err := x509.MarshalPKIXPublicKey(j.publicKey)
	if err != nil {
		return ""
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}
	return string(pem.EncodeToMemory(publicKeyBlock))
}

// GenerateRSAKeys generates RSA key pair for JWT signing
func GenerateRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// LoadRSAPrivateKeyFromPEM loads RSA private key from PEM string
func LoadRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// LoadRSAPublicKeyFromPEM loads RSA public key from PEM string
func LoadRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}