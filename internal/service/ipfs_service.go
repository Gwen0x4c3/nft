package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// IPFSService interface for decentralized storage operations
type IPFSService interface {
	// File operations
	UploadFile(ctx context.Context, file io.Reader, filename string) (string, error)
	DownloadFile(ctx context.Context, hash string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, hash string) error
	
	// JSON/Metadata operations
	UploadJSON(ctx context.Context, data interface{}) (string, error)
	DownloadJSON(ctx context.Context, hash string, dest interface{}) error
	
	// Directory operations
	UploadDirectory(ctx context.Context, files map[string]io.Reader) (string, error)
	ListDirectoryContents(ctx context.Context, hash string) ([]IPFSFile, error)
	
	// Pin management
	PinFile(ctx context.Context, hash string) error
	UnpinFile(ctx context.Context, hash string) error
	ListPinnedFiles(ctx context.Context) ([]IPFSPin, error)
	
	// Status and health
	IsOnline() bool
	GetNodeID() string
	GetStorageStats(ctx context.Context) (*IPFSStats, error)
}

// IPFSFile represents a file in IPFS
type IPFSFile struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size uint64 `json:"size"`
	Type string `json:"type"` // "file" or "directory"
}

// IPFSPin represents a pinned file in IPFS
type IPFSPin struct {
	Hash      string    `json:"hash"`
	Type      string    `json:"type"`
	Size      uint64    `json:"size"`
	PinnedAt  time.Time `json:"pinned_at"`
	Recursive bool      `json:"recursive"`
}

// IPFSStats represents IPFS node statistics
type IPFSStats struct {
	RepoSize      uint64 `json:"repo_size"`
	StorageMax    uint64 `json:"storage_max"`
	NumObjects    uint64 `json:"num_objects"`
	RepoPath      string `json:"repo_path"`
	Version       string `json:"version"`
	PeerID        string `json:"peer_id"`
	ConnectedPeers int   `json:"connected_peers"`
}

// IPFSError represents IPFS-specific errors
type IPFSError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Hash    string `json:"hash,omitempty"`
}

func (e IPFSError) Error() string {
	return e.Message
}

// Common IPFS error codes
const (
	ErrCodeHashNotFound     = 5001
	ErrCodeUploadFailed     = 5002
	ErrCodeDownloadFailed   = 5003
	ErrCodeInvalidHash      = 5004
	ErrCodeNodeOffline      = 5005
	ErrCodeInsufficientSpace = 5006
	ErrCodePinFailed        = 5007
	ErrCodeUnpinFailed      = 5008
)

// IPFSConfig represents IPFS service configuration
type IPFSConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	GatewayURL  string `json:"gateway_url"`
	NodeID      string `json:"node_id"`
	AutoPin     bool   `json:"auto_pin"`
	Timeout     int    `json:"timeout_seconds"`
}

// MockIPFSService provides a mock implementation for testing
type MockIPFSService struct {
	online    bool
	nodeID    string
	files     map[string][]byte
	pins      map[string]IPFSPin
	gatewayURL string
}

// NewMockIPFSService creates a new mock IPFS service
func NewMockIPFSService() *MockIPFSService {
	return &MockIPFSService{
		online:     true,
		nodeID:     "12D3KooWGBQ" + generateRandomHex(32), // Mock peer ID
		files:      make(map[string][]byte),
		pins:       make(map[string]IPFSPin),
		gatewayURL: "https://ipfs.io/ipfs/",
	}
}

// Implementation of IPFSService interface for mock
func (m *MockIPFSService) UploadFile(ctx context.Context, file io.Reader, filename string) (string, error) {
	if !m.online {
		return "", IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", IPFSError{Code: ErrCodeUploadFailed, Message: fmt.Sprintf("failed to read file: %v", err)}
	}

	// Generate mock IPFS hash
	hash := generateIPFSHash(content)
	
	// Store file content
	m.files[hash] = content
	
	// Auto-pin if enabled
	m.pins[hash] = IPFSPin{
		Hash:      hash,
		Type:      "file",
		Size:      uint64(len(content)),
		PinnedAt:  time.Now(),
		Recursive: false,
	}
	
	return hash, nil
}

func (m *MockIPFSService) DownloadFile(ctx context.Context, hash string) (io.ReadCloser, error) {
	if !m.online {
		return nil, IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	content, exists := m.files[hash]
	if !exists {
		return nil, IPFSError{Code: ErrCodeHashNotFound, Message: fmt.Sprintf("file not found: %s", hash)}
	}

	return io.NopCloser(strings.NewReader(string(content))), nil
}

func (m *MockIPFSService) DeleteFile(ctx context.Context, hash string) error {
	if !m.online {
		return IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	delete(m.files, hash)
	delete(m.pins, hash)
	return nil
}

func (m *MockIPFSService) UploadJSON(ctx context.Context, data interface{}) (string, error) {
	if !m.online {
		return "", IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", IPFSError{Code: ErrCodeUploadFailed, Message: fmt.Sprintf("failed to marshal JSON: %v", err)}
	}

	return m.UploadFile(ctx, strings.NewReader(string(jsonData)), "metadata.json")
}

func (m *MockIPFSService) DownloadJSON(ctx context.Context, hash string, dest interface{}) error {
	if !m.online {
		return IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	reader, err := m.DownloadFile(ctx, hash)
	if err != nil {
		return err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return IPFSError{Code: ErrCodeDownloadFailed, Message: fmt.Sprintf("failed to read JSON: %v", err)}
	}

	if err := json.Unmarshal(content, dest); err != nil {
		return IPFSError{Code: ErrCodeDownloadFailed, Message: fmt.Sprintf("failed to unmarshal JSON: %v", err)}
	}

	return nil
}

func (m *MockIPFSService) UploadDirectory(ctx context.Context, files map[string]io.Reader) (string, error) {
	if !m.online {
		return "", IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	// For mock implementation, we'll just generate a hash
	hash := "Qm" + generateRandomHex(44) // Mock directory hash
	
	// Store individual files
	for filename, reader := range files {
		content, err := io.ReadAll(reader)
		if err != nil {
			return "", IPFSError{Code: ErrCodeUploadFailed, Message: fmt.Sprintf("failed to read file %s: %v", filename, err)}
		}
		
		fileHash := generateIPFSHash(content)
		m.files[fileHash] = content
	}
	
	return hash, nil
}

func (m *MockIPFSService) ListDirectoryContents(ctx context.Context, hash string) ([]IPFSFile, error) {
	if !m.online {
		return nil, IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	// Mock directory listing
	return []IPFSFile{
		{Name: "image.jpg", Hash: "Qm" + generateRandomHex(44), Size: 1024000, Type: "file"},
		{Name: "metadata.json", Hash: "Qm" + generateRandomHex(44), Size: 2048, Type: "file"},
	}, nil
}

func (m *MockIPFSService) PinFile(ctx context.Context, hash string) error {
	if !m.online {
		return IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	_, exists := m.files[hash]
	if !exists {
		return IPFSError{Code: ErrCodeHashNotFound, Message: fmt.Sprintf("file not found: %s", hash)}
	}

	m.pins[hash] = IPFSPin{
		Hash:      hash,
		Type:      "file",
		Size:      uint64(len(m.files[hash])),
		PinnedAt:  time.Now(),
		Recursive: false,
	}

	return nil
}

func (m *MockIPFSService) UnpinFile(ctx context.Context, hash string) error {
	if !m.online {
		return IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	delete(m.pins, hash)
	return nil
}

func (m *MockIPFSService) ListPinnedFiles(ctx context.Context) ([]IPFSPin, error) {
	if !m.online {
		return nil, IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	pins := make([]IPFSPin, 0, len(m.pins))
	for _, pin := range m.pins {
		pins = append(pins, pin)
	}

	return pins, nil
}

func (m *MockIPFSService) IsOnline() bool {
	return m.online
}

func (m *MockIPFSService) GetNodeID() string {
	return m.nodeID
}

func (m *MockIPFSService) GetStorageStats(ctx context.Context) (*IPFSStats, error) {
	if !m.online {
		return nil, IPFSError{Code: ErrCodeNodeOffline, Message: "IPFS node is offline"}
	}

	totalSize := uint64(0)
	for _, content := range m.files {
		totalSize += uint64(len(content))
	}

	return &IPFSStats{
		RepoSize:       totalSize,
		StorageMax:     10000000000, // 10GB
		NumObjects:     uint64(len(m.files)),
		RepoPath:       "/mock/ipfs",
		Version:        "0.12.0",
		PeerID:         m.nodeID,
		ConnectedPeers: 25,
	}, nil
}

// Helper functions
func generateIPFSHash(content []byte) string {
	// Generate a mock IPFS hash (Qm...)
	hash := "Qm" + generateRandomHex(44)
	return hash
}


// Utility functions for IPFS operations

// GetIPFSURL constructs a full IPFS URL from hash
func GetIPFSURL(gatewayURL, hash string) string {
	return gatewayURL + hash
}

// ValidateIPFSHash checks if a string is a valid IPFS hash
func ValidateIPFSHash(hash string) bool {
	if len(hash) < 46 {
		return false
	}
	
	// Basic validation for CIDv0 (Qm...) and CIDv1 (b...)
	return strings.HasPrefix(hash, "Qm") || strings.HasPrefix(hash, "b")
}

// ExtractHashFromURL extracts IPFS hash from gateway URL
func ExtractHashFromURL(url string) string {
	// Handle various IPFS gateway URL formats
	if idx := strings.LastIndex(url, "/ipfs/"); idx != -1 {
		return url[idx+6:]
	}
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[idx+1:]
	}
	return url
}

// GetMIMEType determines MIME type from filename
func GetMIMEType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".json":
		return "application/json"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}