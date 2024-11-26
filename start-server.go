package main

import (
	"fmt"
	"log"
	"net"
)

func startServer(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on the port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		mu.Lock()
		if len(clientsList) >= maximumConnections {
			mu.Unlock()
			conn.Write([]byte("Chat is full. Please try again later.\n"))
			conn.Close()
			continue
		}
		mu.Unlock()

		go handleNewConnection(conn)
	}
}
