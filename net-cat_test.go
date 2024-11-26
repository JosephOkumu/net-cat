package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockConn implements net.Conn interface for testing
type MockConn struct {
	net.Conn
	buffer    bytes.Buffer
	closed    bool
	writeMu   sync.Mutex
	closeOnce sync.Once
	ReadFunc  func([]byte) (int, error)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	if m.closed {
		return 0, net.ErrClosed
	}
	return m.buffer.Write(b)
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockConn) Close() error {
	m.closeOnce.Do(func() {
		m.closed = true
	})
	return nil
}

func (m *MockConn) GetWrittenData() string {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	return m.buffer.String()
}

// Required methods to implement net.Conn interface
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestBroadcastMessage(t *testing.T) {
	// Reset global variables for testing
	clientsList = make([]*Client, 0)
	messageHistory = make([]string, 0)
	mu = sync.Mutex{}
	historyMu = sync.Mutex{}

	tests := []struct {
		name          string
		message       string
		numClients    int
		failingClient int // Index of client that should fail on write (-1 for none)
		expectHistory bool
	}{
		{
			name:          "Basic broadcast",
			message:       "Hello, everyone!",
			numClients:    3,
			failingClient: -1,
			expectHistory: true,
		},
		{
			name:          "Empty message",
			message:       "",
			numClients:    2,
			failingClient: -1,
			expectHistory: true,
		},
		{
			name:          "Single client",
			message:       "Solo message",
			numClients:    1,
			failingClient: -1,
			expectHistory: true,
		},
		{
			name:          "Failing client",
			message:       "Test with failing client",
			numClients:    3,
			failingClient: 1,
			expectHistory: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear global state
			clientsList = make([]*Client, 0)
			messageHistory = make([]string, 0)

			// Create mock connections and clients
			mockConns := make([]*MockConn, tt.numClients)
			for i := 0; i < tt.numClients; i++ {
				mockConns[i] = &MockConn{}
				client := &Client{
					conn: mockConns[i],
					name: fmt.Sprintf("TestClient%d", i),
				}
				clientsList = append(clientsList, client)

				// Make one client fail if specified
				if i == tt.failingClient {
					mockConns[i].Close()
				}
			}

			// Create sender connection (separate from other clients)
			senderConn := &MockConn{}

			// Broadcast message
			broadcastMessage(tt.message, senderConn)

			// Give removeClient goroutine time to execute
			time.Sleep(100 * time.Millisecond)

			// Verify message history
			if tt.expectHistory {
				historyMu.Lock()
				if len(messageHistory) != 1 || messageHistory[0] != tt.message {
					t.Errorf("Message not properly stored in history. Expected %q, got %v",
						tt.message, messageHistory)
				}
				historyMu.Unlock()
			}

			// Verify client list state
			mu.Lock()
			currentClients := len(clientsList)
			expectedClients := tt.numClients
			if tt.failingClient >= 0 {
				expectedClients--
			}
			if currentClients != expectedClients {
				t.Errorf("Expected %d clients, but got %d", expectedClients, currentClients)
			}

			// Verify the failing client was removed
			if tt.failingClient >= 0 {
				failedConnFound := false
				for _, client := range clientsList {
					if client.conn == mockConns[tt.failingClient] {
						failedConnFound = true
						break
					}
				}
				if failedConnFound {
					t.Error("Failed client should have been removed from clientsList")
				}
			}
			mu.Unlock()

			// Verify message delivery
			for i, mockConn := range mockConns {
				if i == tt.failingClient {
					continue
				}
				receivedMsg := mockConn.GetWrittenData()
				if receivedMsg != tt.message {
					t.Errorf("Client %d received incorrect message. Expected %q, got %q",
						i, tt.message, receivedMsg)
				}
			}

			// Verify sender didn't receive the message
			senderData := senderConn.GetWrittenData()
			if senderData != "" {
				t.Errorf("Sender received their own message: %q", senderData)
			}
		})
	}
}

func TestDisplayPreviousMessages(t *testing.T) {
	tests := []struct {
		name           string
		messageHistory []string
		expectHeader   bool
		expectFooter   bool
	}{
		{
			name:           "Empty history",
			messageHistory: []string{},
			expectHeader:   false,
			expectFooter:   false,
		},
		{
			name:           "Single message",
			messageHistory: []string{"User1: Hello\n"},
			expectHeader:   true,
			expectFooter:   true,
		},
		{
			name: "Multiple messages",
			messageHistory: []string{
				"User1: First message\n",
				"User2: Second message\n",
				"User3: Third message\n",
			},
			expectHeader: true,
			expectFooter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global message history
			messageHistory = make([]string, 0)
			historyMu = sync.Mutex{}

			// Setup message history
			messageHistory = append(messageHistory, tt.messageHistory...)

			// Create mock connection and client
			mockConn := &MockConn{}
			client := &Client{
				conn: mockConn,
				name: "TestClient",
			}

			// Display previous messages
			displayPreviousMessages(client)

			// Get the output
			output := mockConn.GetWrittenData()

			// Verify header presence
			hasHeader := strings.Contains(output, "\n Previous Messages \n")
			if hasHeader != tt.expectHeader {
				t.Errorf("Header presence mismatch. Expected: %v, Got: %v",
					tt.expectHeader, hasHeader)
			}

			// Verify footer presence
			hasFooter := strings.Contains(output, " End Of Previous Messages \n\n")
			if hasFooter != tt.expectFooter {
				t.Errorf("Footer presence mismatch. Expected: %v, Got: %v",
					tt.expectFooter, hasFooter)
			}

			// Verify all messages are present
			for _, expectedMsg := range tt.messageHistory {
				if !strings.Contains(output, expectedMsg) {
					t.Errorf("Expected message not found in output: %s", expectedMsg)
				}
			}

			// Verify correct order of messages
			if len(tt.messageHistory) > 0 {
				// Remove header and footer for message order verification
				messagesSection := output
				if tt.expectHeader {
					messagesSection = strings.TrimPrefix(messagesSection,
						"\n Previous Messages \n")
				}
				if tt.expectFooter {
					messagesSection = strings.TrimSuffix(messagesSection,
						" End Of Previous Messages \n\n")
				}

				// Split into individual messages
				receivedMessages := strings.Split(strings.TrimSpace(messagesSection), "\n")

				// Verify each message matches and is in correct order
				if len(tt.messageHistory) > 0 {
					for i, expectedMsg := range tt.messageHistory {
						expectedMsg = strings.TrimSpace(expectedMsg)
						receivedMsg := strings.TrimSpace(receivedMessages[i])
						if expectedMsg != receivedMsg {
							t.Errorf("Message at position %d mismatch.\nExpected: %q\nGot: %q",
								i, expectedMsg, receivedMsg)
						}
					}
				}
			}
		})
	}
}

func TestRemoveClient(t *testing.T) {
	// Setup: Create mock connections and add clients to clientsList
	clientsList = make([]*Client, 0)

	mockConn1 := &MockConn{}
	mockConn2 := &MockConn{}
	mockConn3 := &MockConn{}

	clientsList = append(clientsList, &Client{conn: mockConn1, name: "Client1"})
	clientsList = append(clientsList, &Client{conn: mockConn2, name: "Client2"})
	clientsList = append(clientsList, &Client{conn: mockConn3, name: "Client3"})

	// Ensure the initial list has 3 clients
	if len(clientsList) != 3 {
		t.Errorf("Expected 3 clients, but got %d", len(clientsList))
	}

	// Test: Remove second client (mockConn2)
	removeClient(mockConn2)

	// Verify: List should now contain 2 clients
	if len(clientsList) != 2 {
		t.Errorf("Expected 2 clients after removal, but got %d", len(clientsList))
	}

	// Verify: Ensure mockConn2 (Client2) is removed
	clientFound := false
	for _, client := range clientsList {
		if client.conn == mockConn2 {
			clientFound = true
			break
		}
	}

	if clientFound {
		t.Error("Client with mockConn2 was not removed from clientsList")
	}

	// Verify: Ensure mockConn2 is closed
	if !mockConn2.closed {
		t.Error("mockConn2 was not closed during removal")
	}

	// Test: Remove the first client (mockConn1)
	removeClient(mockConn1)

	// Verify: List should now contain 1 client
	if len(clientsList) != 1 {
		t.Errorf("Expected 1 client after removing the first client, but got %d", len(clientsList))
	}

	// Verify: Ensure mockConn1 (Client1) is removed
	clientFound = false
	for _, client := range clientsList {
		if client.conn == mockConn1 {
			clientFound = true
			break
		}
	}

	if clientFound {
		t.Error("Client with mockConn1 was not removed from clientsList")
	}

	// Verify: Ensure mockConn1 is closed
	if !mockConn1.closed {
		t.Error("mockConn1 was not closed during removal")
	}

	// Test: Remove the last client (mockConn3)
	removeClient(mockConn3)

	// Verify: List should now be empty
	if len(clientsList) != 0 {
		t.Errorf("Expected 0 clients after removing the last client, but got %d", len(clientsList))
	}

	// Verify: Ensure mockConn3 (Client3) is closed
	if !mockConn3.closed {
		t.Error("mockConn3 was not closed during removal")
	}
}

func TestStartServer(t *testing.T) {
	// Create a channel to signal server shutdown
	shutdown := make(chan bool)
	testPort := ":8080"

	// Start server in a goroutine
	go func() {
		defer close(shutdown)
		startServer(testPort)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	t.Run("Basic Connection with Name", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost"+testPort)
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		// Read the welcome message
		reader := bufio.NewReader(conn)
		_, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read welcome message: %v", err)
		}

		// Send client name
		name := "TestUser\n"
		_, err = conn.Write([]byte(name))
		if err != nil {
			t.Fatalf("Failed to send name: %v", err)
		}

		// Wait briefly for server to process
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		found := false
		for _, client := range clientsList {
			if client.name == "TestUser" {
				found = true
				break
			}
		}
		mu.Unlock()

		if !found {
			t.Error("Client was not properly added to clientsList")
		}
	})

	t.Run("Exceed Maximum Connections", func(t *testing.T) {
		var conns []net.Conn
		defer func() {
			for _, conn := range conns {
				conn.Close()
			}
		}()

		// Create maximum allowed connections
		for i := 0; i < maximumConnections; i++ {
			conn, err := net.Dial("tcp", "localhost"+testPort)
			if err != nil {
				t.Fatalf("Failed to connect to server: %v", err)
			}
			conns = append(conns, conn)

			// Send client name
			_, err = conn.Write([]byte(fmt.Sprintf("TestUser%d\n", i)))
			if err != nil {
				t.Fatalf("Failed to send name for client %d: %v", i, err)
			}

			// Wait for the server to process the connection
			time.Sleep(50 * time.Millisecond)
		}

		// Attempt to connect one more client beyond the limit
		exceedConn, err := net.Dial("tcp", "localhost"+testPort)
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer exceedConn.Close()

		// Read the response for the exceeded connection
		reader := bufio.NewReader(exceedConn)
		message, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response from server: %v", err)
		}

		// Check that the response indicates the chat is full
		expectedMessage := "Chat is full. Please try again later.\n"
		if message != expectedMessage {
			t.Errorf("Expected message %q, got %q", expectedMessage, message)
		}

		// Verify that the exceeded connection was not added to clientsList
		mu.Lock()
		defer mu.Unlock()
		if len(clientsList) > maximumConnections {
			t.Errorf("Exceeded maximum connections limit: expected %d, got %d", maximumConnections, len(clientsList))
		}
	})
}
