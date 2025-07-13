package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read the request line (first line of the HTTP request)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err.Error())
		return
	}

	// Example request line: GET / HTTP/1.1\r\n
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) < 2 {
		fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
		return
	}

	path := parts[1]

	// Route based on path
	if path == "/" {
		fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n\r\n")
	} else {
		fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}
