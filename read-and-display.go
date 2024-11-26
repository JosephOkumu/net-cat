package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

// Will open a specified txt file to read its contents.
func readNetcatLogo(fileName string) (logo string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		logo += scanner.Text() + "\n"
	}
	return
}

// Uses the Write property to display the information in the specified file.
func displayLogo(conn net.Conn) {
	conn.Write([]byte(readNetcatLogo("net-cat-logo.txt")))
}
