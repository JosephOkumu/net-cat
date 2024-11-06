package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultPort = ":8989"
	welcomeBanner = `Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    .       | ' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     -'       --'
[ENTER YOUR NAME]: `
)

// Client struct represents a connected client
type Client struct {
	conn net.Conn
	name string
}

var (
	mu                 sync.Mutex
	maximumConnections = 10
	clientsList        = make([]*Client, 0, maximumConnections)
)

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

func handleNewConnection(conn net.Conn) {
	conn.Write([]byte(welcomeBanner))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading client name: %v", err)
		conn.Close()
		return
	}

	name := strings.TrimSpace(string(buffer[:n]))
	if name == "" {
		conn.Write([]byte("Name cannot be empty\n"))
		conn.Close()
		return
	}

	newClient := &Client{
		conn: conn,
		name: name,
	}

	mu.Lock()
	clientsList = append(clientsList, newClient)
	mu.Unlock()

	broadcastMessage(fmt.Sprintf("\n%s has joined our chat...\n", name), conn)

	handleClient(newClient)
}

func broadcastMessage(message string, senderConn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for _, client := range clientsList {
		if client.conn != senderConn {
			_, err := client.conn.Write([]byte(message))
			if err != nil {
				log.Printf("Error broadcasting to %s: %v", client.name, err)
				removeClient(client.conn)
			}
		}
	}
}

func removeClient(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for i, client := range clientsList {
		if client.conn == conn {
			client.conn.Close()
			broadcastMessage(fmt.Sprintf("\n%s has left our chat...\n", client.name), client.conn)
			clientsList = append(clientsList[:i], clientsList[i+1:]...)
			break
		}
	}
}

func findClientByConn(conn net.Conn) *Client {
	mu.Lock()
	defer mu.Unlock()
	
	for _, client := range clientsList {
		if client.conn == conn {
			return client
		}
	}
	return nil
}

// handleClient function will be implemented to handle individual client messages
func handleClient(client *Client) {
	// This function will be implemented to handle client message processing
}