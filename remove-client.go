package main

import (
	"net"
)

// Removes a client from the server upon termination from the client.
func removeClient(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for i, client := range clientsList {
		if client.conn == conn {
			client.conn.Close()
			
			// Remove client from the list without broadcasting
			clientsList = append(clientsList[:i], clientsList[i+1:]...)
			break
		}
	}
}
