package websocket

// This file provides minimal event broadcasting helpers (Task T085).
// For now it's a thin wrapper around Manager methods to keep code organized.

// Predefined event names based on contract tests & specification
const (
	EventBidPlaced           = "bid_placed"
	EventAuctionStatusChange = "auction_status_changed"
	EventNFTTransferred      = "nft_transferred"
	EventNotification        = "notification"
	EventSubscription        = "subscription_confirmed"
	EventPong                = "pong"
)

// ManagerEventAPI exposes higher-level broadcasting helpers.
// In a fuller implementation these would enrich payloads, perform filtering, etc.
func (m *Manager) BroadcastBidPlaced(data interface{}) { m.BroadcastEvent(EventBidPlaced, data) }
func (m *Manager) BroadcastAuctionStatusChanged(data interface{}) { m.BroadcastEvent(EventAuctionStatusChange, data) }
func (m *Manager) BroadcastNFTTransferred(data interface{}) { m.BroadcastEvent(EventNFTTransferred, data) }
func (m *Manager) BroadcastNotification(data interface{}) { m.BroadcastEvent(EventNotification, data) }
