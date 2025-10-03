package types

import "time"

// Transaction represents a blockchain transaction
type Transaction struct {
	Hash   string    `json:"hash"`
	Status string    `json:"status"`
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount string    `json:"amount"`
	Gas    int64     `json:"gas"`
	Nonce  int       `json:"nonce"`
	Time   time.Time `json:"time"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
