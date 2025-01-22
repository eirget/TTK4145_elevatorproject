package main

import (
	"fmt"
	"net"
	"time"
)

func udpListen(port int) {
	// Create a UDP address for the listener.
	address := fmt.Sprintf(":%d", port)
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Set up a UDP connection to listen on the port.
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error setting up listener:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Listening for broadcasts on port %d...\n", port)

	// Create a buffer to read incoming data.
	buffer := make([]byte, 1024)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // Set a timeout

	// Read data from the connection.
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error or timeout reading from connection:", err)
		return
	}

	// Print the received message and sender address
	fmt.Printf("Received message from %s: %s\n", addr, string(buffer[:n]))

	go udpSend("Hello :)", addr.IP.String(), 20022)
}



func udpSend(message string, address string, port int) {
	// Resolve the UDP address of the destination
	destAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, destAddr)
	if err != nil {
		fmt.Println("Error creating connection:", err)
		return
	}
	defer conn.Close() // Ensure the connection is closed after sending the message

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	fmt.Printf("Message sent to %s:%d: %s\n", address, port, message)
}

func listenForResponses(port int, timeout time.Duration) {
	// Resolve the UDP address for listening on the given port
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Set up the UDP listener
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error setting up listener:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Listening for responses on port %d...\n", port)

	// Buffer to hold incoming data
	buffer := make([]byte, 1024)

	// Wait for and read the response with timeout
	conn.SetReadDeadline(time.Now().Add(timeout)) // Set a timeout

	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error or timeout reading from connection:", err)
		return
	}

	// Print the received message and sender address
	fmt.Printf("Received message from %s: %s\n", addr, string(buffer[:n]))
}


func main() {

	//server_IP := "10.100.23.204" //find IP for server

	//udpSend("Hello :)", server_IP, 20022)
	udpListen(30000)
	listenForResponses(20022, 5*time.Second)
}
