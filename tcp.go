package main

import (
	"fmt"
	"net"

	//"os"
	"time"
)




func tcpServer(IP string, port int) {
	// Create a TCP address for the listener.
	address := fmt.Sprintf("%s:%d", IP, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Set up a TCP connection to listen on the port.
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Error setting up listener:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Listening for TCP connections on port %d...\n", port)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create a buffer to read incoming data.
	buffer := make([]byte, 1024)

	// Read data from the connection
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}

		// Print the received message
		fmt.Printf("Received message: %s\n", string(buffer[:n]))
	}

	
}

func tcpClient(address string, port int) {
	// Resolve the TCP address of the destination
	destAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Create a TCP connection
	conn, err := net.DialTCP("tcp", nil, destAddr)
	if err != nil {
		fmt.Println("Error creating connection:", err)
		return
	}

	go handleConnection(conn)

	time.Sleep(1 * time.Second)
	conn.Close() // Ensure the connection is closed after sending the message

	
}

func main() {
	addr := "10.100.23.204"

	tcpClient(addr, 33546)

	select {}
}
