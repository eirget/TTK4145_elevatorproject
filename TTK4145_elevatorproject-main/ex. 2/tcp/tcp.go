package main

import (
	"fmt"
	"net"
	//"os"
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
		//print message
		message := string(buffer[:n])
		fmt.Printf("Recieved message: %s\n", message)

		//echo message back
		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Error writing to connection:", err)
			return
		}
	}

}

func tcpClient(serverIP string, serverPort int, localPort int) {
	// Resolve the TCP address of the destination
	destAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Create a TCP connection to server
	conn, err := net.DialTCP("tcp", nil, destAddr)
	if err != nil {
		fmt.Println("Error creating connection:", err)
		return
	}

	defer conn.Close()
	//go handleConnection(conn) ikke riktig

	//send the "connect to" message
	localIP, err := getLocalIP()
	if err != nil {
		fmt.Printf("Error determining local IP:", err)
		return
	}

	connectMessage := fmt.Sprintf("Connect to: %s:%d\000", localIP, localPort)
	_, err = conn.Write([]byte(connectMessage))
	if err != nil {
		fmt.Println("Error sending connect message:", err)
		return
	}

	//handle messages from server
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from server:", err)
			return
		}
		fmt.Printf("Server: %s\n", string(buffer[:n]))
	}

	//time.Sleep(1 * time.Second)
	//conn.Close() // Ensure the connection is closed after sending the message

}

func getLocalIP() (string, error) {
	address, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			return ipNet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no IP address found")
}

func main() {
	addr := "10.100.23.204"

	go tcpServer("0.0.0.0", 20022)

	tcpClient(addr, 33546, 20022) //this line is the connection part
	select {}
}
