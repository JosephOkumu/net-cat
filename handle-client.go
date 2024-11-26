package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Will handle individual client message processing and broadcasting.
func handleClient(client *Client) {
	defer removeClient(client.conn)
	buffer := make([]byte, 4096)

	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			// Check for EOF (end of file)
			if err.Error() == "EOF" {
				// Notify others about client leaving before removing them
				broadcastMessage(fmt.Sprintf("\n%s has left our chat...\n", client.name), client.conn)
				return
			}
			// Handle other errors
			log.Printf("Error reading from %s: %v\n", client.name, err)
			return
		}

		message := strings.TrimSpace(string(buffer[:n]))
		if message == "" {
			log.Println("Cannot send an empty message to the chat.")
			break
		}
		// Broadcast the message
		formattedMessage := fmt.Sprintf("[%s][%s]: %s\n", time.Now().Format(time.DateTime), client.name, message)
		broadcastMessage(formattedMessage, client.conn)
	}
}


