package main

import (
	"log"
	"net"
)

// Will always send out broadcast messages to everyone on the chat room except the sender of the message.
func broadcastMessage(message string, senderConn net.Conn) {
	// Store messages in history first
	historyMu.Lock()
	messageHistory = append(messageHistory, message)
	historyMu.Unlock()
	
	mu.Lock()
	defer mu.Unlock()

	for _, client := range clientsList {
		// Exclude the message sender from seeing their triggered broadcast event.
		if client.conn != senderConn {
			_, err := client.conn.Write([]byte(message))
			if err != nil {
				log.Printf("Error broadcasting to %v: %v\n", &client.name, err)
				go removeClient(client.conn)
			}
		}
	}
}
