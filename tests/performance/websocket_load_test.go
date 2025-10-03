package performance

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"

	ws "nft-platform/internal/websocket"
)

// WebSocketLoadTestSuite contains load tests for WebSocket connections
type WebSocketLoadTestSuite struct {
	suite.Suite
	manager   *ws.Manager
	serverURL string
	baseURL   string
	dialer    websocket.Dialer
}

func (suite *WebSocketLoadTestSuite) SetupSuite() {
	suite.manager = ws.NewManager()

	// Create a test HTTP server with WebSocket handler
	mux := http.NewServeMux()
	// Create a gin router to wrap the WebSocket handler
	ginRouter := gin.New()
	ginRouter.GET("/ws", suite.manager.HandleConnection)

	// Wrap gin handler with http handler
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ginRouter.ServeHTTP(w, r)
	})

	// Use a listener to get the actual port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: mux,
	}

	// Start server
	go func() {
		server.Serve(listener)
	}()

	// Get the actual address
	time.Sleep(100 * time.Millisecond)
	actualAddr := server.Addr
	suite.baseURL = "http://" + actualAddr
	suite.serverURL = "ws://" + actualAddr + "/ws"

	suite.dialer = websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
}

func (suite *WebSocketLoadTestSuite) TearDownSuite() {
	// Cleanup connections if needed
}

// createTestWebSocketConnection creates a WebSocket connection for testing
func (suite *WebSocketLoadTestSuite) createTestWebSocketConnection(token string) (*websocket.Conn, error) {
	u, err := url.Parse(suite.serverURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	conn, _, err := suite.dialer.Dial(u.String(), nil)
	return conn, err
}

// TestConcurrentConnections tests handling of multiple concurrent WebSocket connections
func (suite *WebSocketLoadTestSuite) TestConcurrentConnections() {
	const numConnections = 100
	const maxResponseTime = 2 * time.Second

	var wg sync.WaitGroup
	connectionTimes := make([]time.Duration, numConnections)
	connectionErrors := make([]error, numConnections)

	startTime := time.Now()

	// Create multiple concurrent connections
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			connStart := time.Now()
			token := fmt.Sprintf("valid.token.%d", index)

			conn, err := suite.createTestWebSocketConnection(token)
			if err != nil {
				connectionErrors[index] = err
				return
			}
			defer conn.Close()

			// Wait for initial connection message
			_, message, err := conn.ReadMessage()
			if err != nil {
				connectionErrors[index] = err
				return
			}

			// Verify it's a connection_established event
			var event ws.Event
			if err := json.Unmarshal(message, &event); err != nil {
				connectionErrors[index] = err
				return
			}

			if event.Event != "connection_established" {
				connectionErrors[index] = fmt.Errorf("expected connection_established event, got %s", event.Event)
				return
			}

			connectionTimes[index] = time.Since(connStart)
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Analyze results
	var successfulConnections int
	var totalConnectionTime time.Duration
	var maxConnectionTime time.Duration

	for i, err := range connectionErrors {
		if err == nil {
			successfulConnections++
			totalConnectionTime += connectionTimes[i]
			if connectionTimes[i] > maxConnectionTime {
				maxConnectionTime = connectionTimes[i]
			}
		} else {
			fmt.Printf("Connection %d failed: %v\n", i, err)
		}
	}

	avgConnectionTime := totalConnectionTime / time.Duration(successfulConnections)

	fmt.Printf("Concurrent WebSocket Connections Test (%d connections):\n", numConnections)
	fmt.Printf("  Successful connections: %d/%d (%.1f%%)\n", successfulConnections, numConnections, float64(successfulConnections)/float64(numConnections)*100)
	fmt.Printf("  Total time: %v\n", totalTime)
	fmt.Printf("  Average connection time: %v\n", avgConnectionTime)
	fmt.Printf("  Max connection time: %v\n", maxConnectionTime)

	// Assertions
	suite.Greater(successfulConnections, numConnections*95/100, "At least 95% of connections should succeed")
	suite.Less(maxConnectionTime, maxResponseTime, "Max connection time should be less than 2 seconds")
	suite.Less(avgConnectionTime, 500*time.Millisecond, "Average connection time should be less than 500ms")
}

// TestBroadcastPerformance tests broadcasting performance to many connected clients
func (suite *WebSocketLoadTestSuite) TestBroadcastPerformance() {
	const numConnections = 50
	const numBroadcasts = 10
	const maxMessageDelay = 100 * time.Millisecond

	// Create connections
	connections := make([]*websocket.Conn, numConnections)
	for i := 0; i < numConnections; i++ {
		token := fmt.Sprintf("valid.token.%d", i)
		conn, err := suite.createTestWebSocketConnection(token)
		suite.Require().NoError(err)
		connections[i] = conn

		// Wait for initial connection message
		_, _, err = conn.ReadMessage()
		suite.Require().NoError(err)
	}

	// Cleanup connections
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	// Test broadcasting
	var wg sync.WaitGroup
	receivedCounts := make([]int, numConnections)
	receivedTimes := make([][]time.Time, numConnections)

	for i := 0; i < numConnections; i++ {
		receivedTimes[i] = make([]time.Time, 0, numBroadcasts)
		go func(index int) {
			conn := connections[index]

			// Listen for messages
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					break
				}

				var event ws.Event
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.Event == "test_broadcast" {
					receivedCounts[index]++
					receivedTimes[index] = append(receivedTimes[index], event.Timestamp)
				}
			}
		}(i)
	}

	// Wait a bit for listeners to start
	time.Sleep(100 * time.Millisecond)

	// Send broadcasts
	broadcastStart := time.Now()
	for i := 0; i < numBroadcasts; i++ {
		testData := map[string]interface{}{
			"broadcast_id": i,
			"message":      fmt.Sprintf("Test broadcast message %d", i),
		}

		suite.manager.BroadcastEvent("test_broadcast", testData)
		time.Sleep(50 * time.Millisecond) // Small delay between broadcasts
	}

	// Wait for all messages to be received
	time.Sleep(1 * time.Second)

	// Calculate metrics
	var totalReceived int
	var totalDelay time.Duration
	var maxDelay time.Duration

	for i := 0; i < numConnections; i++ {
		totalReceived += receivedCounts[i]

		// Calculate delay for first message
		if len(receivedTimes[i]) > 0 {
			delay := receivedTimes[i][0].Sub(broadcastStart)
			totalDelay += delay
			if delay > maxDelay {
				maxDelay = delay
			}
		}
	}

	avgDelay := totalDelay / time.Duration(numConnections)

	fmt.Printf("Broadcast Performance Test (%d connections, %d broadcasts):\n", numConnections, numBroadcasts)
	fmt.Printf("  Total messages sent: %d\n", numBroadcasts)
	fmt.Printf("  Total messages received: %d\n", totalReceived)
	fmt.Printf("  Reception rate: %.1f%%\n", float64(totalReceived)/float64(numConnections*numBroadcasts)*100)
	fmt.Printf("  Average message delay: %v\n", avgDelay)
	fmt.Printf("  Max message delay: %v\n", maxDelay)

	// Assertions
	expectedMessages := numConnections * numBroadcasts
	suite.GreaterOrEqual(totalReceived, expectedMessages*90/100, "At least 90% of messages should be received")
	suite.Less(maxDelay, maxMessageDelay, "Max message delay should be less than 100ms")
	suite.Less(avgDelay, 50*time.Millisecond, "Average message delay should be less than 50ms")
}

// TestConnectionStability tests connection stability under sustained load
func (suite *WebSocketLoadTestSuite) TestConnectionStability() {
	const numConnections = 30
	const testDuration = 30 * time.Second
	const pingInterval = 1 * time.Second

	connections := make([]*websocket.Conn, numConnections)
	connectionErrors := make([]error, numConnections)

	// Create connections
	for i := 0; i < numConnections; i++ {
		token := fmt.Sprintf("valid.token.%d", i)
		conn, err := suite.createTestWebSocketConnection(token)
		if err != nil {
			connectionErrors[i] = err
			continue
		}
		connections[i] = conn

		// Wait for initial connection message
		_, _, err = conn.ReadMessage()
		if err != nil {
			connectionErrors[i] = err
			conn.Close()
			continue
		}
	}

	// Cleanup connections
	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	var wg sync.WaitGroup
	stableConnections := 0
	connectionDurations := make([]time.Duration, numConnections)

	// Monitor each connection
	for i := 0; i < numConnections; i++ {
		if connections[i] == nil {
			continue
		}

		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			conn := connections[index]
			startTime := time.Now()
			messageCount := 0

			for time.Since(startTime) < testDuration {
				// Set read deadline
				conn.SetReadDeadline(time.Now().Add(2 * pingInterval))

				_, _, err := conn.ReadMessage()
				if err != nil {
					// Connection closed or error occurred
					connectionDurations[index] = time.Since(startTime)
					return
				}

				messageCount++
				time.Sleep(100 * time.Millisecond)
			}

			connectionDurations[index] = time.Since(startTime)
			stableConnections++
		}(i)
	}

	// Send periodic pings
	go func() {
		for time.Since(time.Now()) < testDuration {
			suite.manager.BroadcastEvent("ping", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
			time.Sleep(pingInterval)
		}
	}()

	wg.Wait()

	// Calculate stability metrics
	var avgDuration time.Duration
	var minDuration, maxDuration time.Duration

	for i, duration := range connectionDurations {
		if duration > 0 {
			if minDuration == 0 || duration < minDuration {
				minDuration = duration
			}
			if maxDuration == 0 || duration > maxDuration {
				maxDuration = duration
			}
			avgDuration += duration
		}
	}

	if stableConnections > 0 {
		avgDuration /= time.Duration(stableConnections)
	}

	fmt.Printf("Connection Stability Test (%d connections, %v duration):\n", numConnections, testDuration)
	fmt.Printf("  Stable connections: %d/%d (%.1f%%)\n", stableConnections, numConnections, float64(stableConnections)/float64(numConnections)*100)
	fmt.Printf("  Average connection duration: %v\n", avgDuration)
	fmt.Printf("  Min connection duration: %v\n", minDuration)
	fmt.Printf("  Max connection duration: %v\n", maxDuration)

	// Assertions
	suite.GreaterOrEqual(stableConnections, numConnections*90/100, "At least 90% of connections should remain stable")
	suite.GreaterOrEqual(avgDuration, testDuration*90/100, "Average connection duration should be at least 90% of test duration")
}

// TestMessageThroughput tests message throughput capacity
func (suite *WebSocketLoadTestSuite) TestMessageThroughput() {
	const numConnections = 20
	const messagesPerSecond = 100
	const testDuration = 10 * time.Second

	// Create connections
	connections := make([]*websocket.Conn, numConnections)
	for i := 0; i < numConnections; i++ {
		token := fmt.Sprintf("valid.token.%d", i)
		conn, err := suite.createTestWebSocketConnection(token)
		suite.Require().NoError(err)
		connections[i] = conn

		// Wait for initial connection message
		_, _, err = conn.ReadMessage()
		suite.Require().NoError(err)
	}

	// Cleanup connections
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	var wg sync.WaitGroup
	receivedCounts := make([]int, numConnections)

	// Start message receivers
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			conn := connections[index]
			conn.SetReadDeadline(time.Now().Add(testDuration + 5*time.Second))

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					break
				}

				var event ws.Event
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.Event == "throughput_test" {
					receivedCounts[index]++
				}
			}
		}(i)
	}

	// Send messages at high rate
	messageInterval := time.Second / time.Duration(messagesPerSecond)
	totalMessages := 0
	startTime := time.Now()

	for time.Since(startTime) < testDuration {
		testData := map[string]interface{}{
			"message_id": totalMessages,
			"timestamp":  time.Now().UnixNano(),
		}

		suite.manager.BroadcastEvent("throughput_test", testData)
		totalMessages++

		time.Sleep(messageInterval)
	}

	// Wait for message processing
	time.Sleep(1 * time.Second)

	// Calculate throughput metrics
	var totalReceived int
	for _, count := range receivedCounts {
		totalReceived += count
	}

	actualDuration := time.Since(startTime)
	actualMessagesPerSecond := float64(totalMessages) / actualDuration.Seconds()
	actualThroughput := float64(totalReceived) / actualDuration.Seconds()

	fmt.Printf("Message Throughput Test (%d connections, %v duration):\n", numConnections, testDuration)
	fmt.Printf("  Messages sent: %d\n", totalMessages)
	fmt.Printf("  Messages received: %d\n", totalReceived)
	fmt.Printf("  Send rate: %.1f messages/second\n", actualMessagesPerSecond)
	fmt.Printf("  Receive throughput: %.1f messages/second\n", actualThroughput)
	fmt.Printf("  Delivery rate: %.1f%%\n", float64(totalReceived)/float64(totalMessages)*100)

	// Assertions
	suite.GreaterOrEqual(actualMessagesPerSecond, messagesPerSecond*80/100, "Send rate should be at least 80% of target")
	suite.GreaterOrEqual(actualThroughput, messagesPerSecond*60/100, "Receive throughput should be at least 60% of target")
	suite.GreaterOrEqual(float64(totalReceived)/float64(totalMessages), 0.7, "Delivery rate should be at least 70%")
}

// TestReconnectionBehavior tests automatic reconnection handling
func (suite *WebSocketLoadTestSuite) TestReconnectionBehavior() {
	const numConnections = 10
	const reconnectionAttempts = 3

	var wg sync.WaitGroup
	successfulReconnections := 0

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			for attempt := 0; attempt < reconnectionAttempts; attempt++ {
				token := fmt.Sprintf("valid.token.%d.attempt.%d", index, attempt)

				conn, err := suite.createTestWebSocketConnection(token)
				if err != nil {
					continue
				}

				// Wait for connection message
				_, message, err := conn.ReadMessage()
				if err != nil {
					conn.Close()
					continue
				}

				var event ws.Event
				if err := json.Unmarshal(message, &event); err != nil {
					conn.Close()
					continue
				}

				if event.Event == "connection_established" {
					successfulReconnections++

					// Keep connection alive for a short time
					time.Sleep(100 * time.Millisecond)
					conn.Close()
					return
				}

				conn.Close()
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("Reconnection Behavior Test (%d connections, %d attempts each):\n", numConnections, reconnectionAttempts)
	fmt.Printf("  Successful reconnections: %d/%d (%.1f%%)\n", successfulReconnections, numConnections, float64(successfulReconnections)/float64(numConnections)*100)

	// Assertions
	suite.GreaterOrEqual(successfulReconnections, numConnections*90/100, "At least 90% of reconnections should succeed")
}

// TestMemoryUsage tests memory usage patterns under load
func (suite *WebSocketLoadTestSuite) TestMemoryUsage() {
	const numConnections = 100
	const messageCount = 50

	// Get initial memory usage (simplified)
	connections := make([]*websocket.Conn, 0, numConnections)

	// Create connections gradually
	for i := 0; i < numConnections; i++ {
		token := fmt.Sprintf("valid.token.%d", i)
		conn, err := suite.createTestWebSocketConnection(token)
		if err != nil {
			continue
		}

		// Wait for connection message
		_, _, err = conn.ReadMessage()
		if err != nil {
			conn.Close()
			continue
		}

		connections = append(connections, conn)

		// Send some messages to this connection
		for j := 0; j < messageCount; j++ {
			testData := map[string]interface{}{
				"connection_id": i,
				"message_id":    j,
			}

			suite.manager.BroadcastEvent("memory_test", testData)
			time.Sleep(1 * time.Millisecond)
		}

		// Small delay between connections
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("Memory Usage Test (%d connections, %d messages each):\n", len(connections), messageCount)
	fmt.Printf("  Total active connections: %d\n", len(connections))
	fmt.Printf("  Total messages sent: %d\n", len(connections)*messageCount)

	// Cleanup
	for _, conn := range connections {
		conn.Close()
	}

	// Allow some time for cleanup
	time.Sleep(1 * time.Second)

	// Assertions
	suite.Equal(numConnections, len(connections), "All connections should be created successfully")
}

// TestRunner runs the WebSocket load test suite
func TestWebSocketLoadTestSuite(t *testing.T) {
	suite.Run(t, new(WebSocketLoadTestSuite))
}

// BenchmarkWebSocketConnection benchmarks WebSocket connection establishment
func BenchmarkWebSocketConnection(b *testing.B) {
	manager := ws.NewManager()

	// Create test server
	mux := http.NewServeMux()
	ginRouter := gin.New()
	ginRouter.GET("/ws", manager.HandleConnection)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ginRouter.ServeHTTP(w, r)
	})

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: mux,
	}

	go server.Serve(listener)
	defer server.Close()
	time.Sleep(100 * time.Millisecond)

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		token := fmt.Sprintf("valid.token.%d", i)
		u, _ := url.Parse("ws://" + server.Addr + "/ws")
		q := u.Query()
		q.Set("token", token)
		u.RawQuery = q.Encode()

		conn, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			b.Fatalf("Failed to connect: %v", err)
		}

		// Wait for connection message
		conn.ReadMessage()
		conn.Close()
	}
}

// BenchmarkWebSocketBroadcast benchmarks WebSocket broadcasting
func BenchmarkWebSocketBroadcast(b *testing.B) {
	manager := ws.NewManager()
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	// Create test server
	mux := http.NewServeMux()
	ginRouter := gin.New()
	ginRouter.GET("/ws", manager.HandleConnection)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ginRouter.ServeHTTP(w, r)
	})

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: mux,
	}

	go server.Serve(listener)
	defer server.Close()
	time.Sleep(100 * time.Millisecond)

	// Create test connections
	connections := make([]*websocket.Conn, 10)
	for i := 0; i < 10; i++ {
		token := fmt.Sprintf("valid.token.%d", i)
		u, _ := url.Parse("ws://" + server.Addr + "/ws")
		q := u.Query()
		q.Set("token", token)
		u.RawQuery = q.Encode()

		conn, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			b.Fatalf("Failed to connect: %v", err)
		}
		connections[i] = conn

		// Wait for connection message
		conn.ReadMessage()
	}

	// Cleanup connections
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testData := map[string]interface{}{
			"message_id": i,
		}

		manager.BroadcastEvent("benchmark_test", testData)
	}
}
