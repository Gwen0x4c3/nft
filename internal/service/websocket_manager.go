package service

// WebSocketManager interface for real-time notification delivery
type WebSocketManager interface {
	SendToUser(userID uint, event string, data interface{}) error
	BroadcastToUsers(userIDs []uint, event string, data interface{}) error
	IsUserConnected(userID uint) bool
	GetConnectedUsers() []uint
}
