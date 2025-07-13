package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Server running on port 4221...")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	// Parse the request line
	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
		return
	}

	path := parts[1]

	if path == "/" {
		// Root route
		fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n\r\n")
	} else if strings.HasPrefix(path, "/echo/") {
		// Echo route
		toEcho := strings.TrimPrefix(path, "/echo/")
		contentLength := len(toEcho)

		response := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
			contentLength, toEcho,
		)

		fmt.Fprint(conn, response)
	} else {
		// Not found
		fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}
