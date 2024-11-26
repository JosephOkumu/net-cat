package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

// Client struct represents a connected client
type Client struct {
	conn net.Conn
	name string
}

var (
	// clientsList holds all connected clients
	clientsList = make([]*Client, 0, maximumConnections)

	// mu is used to synchronize access to shared resources
	mu sync.Mutex

	// maximumConnections defines the maximum number of allowed clients
	maximumConnections = 10

	// Stores all chat messages
	messageHistory []string

	// Mutex for message history
    historyMu sync.Mutex
)

const defaultPort = ":8989"

func main() {
	switch len(os.Args) {
	case 1:
		startServer(defaultPort)
	case 2:
		port := os.Args[1]
		portNum, err := strconv.Atoi(port)
		if err != nil {
			log.Println("Invalid port number: ", err)
			return
		}
		if portNum < 0 || portNum > 65535 {
			log.Println("Invalid port number: <0 or >65535")
			return
		}
		startServer(":" + port)
	default:
		log.Println("[USAGE]: ./TCPChat $port")
		return
	}
}
