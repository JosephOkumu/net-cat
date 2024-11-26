package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Will be used for setting up a chat room environment when there are clients to connect
func handleNewConnection(conn net.Conn) {
	// Have a 1 minute timeout for name input to prevent memory wastage
	conn.SetDeadline(time.Now().Add(60 * time.Second))

	// Invoke function that reads from a text file and displays the net-cat logo
	displayLogo(conn)

	// Prompt username
	conn.Write([]byte("[ENTER YOUR NAME]: "))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading client name: %v\n", err)
		conn.Close()
		return
	}

	// Will reset deadline after getting the name
	conn.SetDeadline(time.Time{})

	name := strings.TrimSpace(string(buffer[:n]))
	if name == "" {
		conn.Write([]byte("Name cannot be empty\n"))
		conn.Close()
		return
	}

	// Check for duplicate names before a client can get connected to the server's chat room
	mu.Lock()
	for _, client := range clientsList {
		if strings.EqualFold(client.name, name) {
			mu.Unlock()
			conn.Write([]byte("This name is already taken. Please try another name to rejoin the chat.\n"))
			conn.Close()
			return
		}
	}
	mu.Unlock()

	newClient := &Client{
		conn: conn,
		name: name,
	}
	mu.Lock()
	clientsList = append(clientsList, newClient)
	mu.Unlock()

	// Will display a welcome message to the new client
	conn.Write([]byte(fmt.Sprintf("Welcome to the chat, %s!\n", name)))

	// Will display the previous messages to the new client
	displayPreviousMessages(newClient)

	// Notify others about the new client
	broadcastMessage(fmt.Sprintf("\n%s has joined our chat...\n", name), conn)

	// Start processing messages for the new client
	go handleClient(newClient)
}
