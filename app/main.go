package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var directory string

func main() {
	// Parse --directory flag
	flag.StringVar(&directory, "directory", ".", "Directory to serve files from")
	flag.Parse()

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

	// Step 1: Read request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	parts := strings.Fields(requestLine)
	if len(parts) < 3 {
		fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
		return
	}
	method := parts[0]
	path := parts[1]

	// Step 2: Parse headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		kv := strings.SplitN(line, ":", 2)
		if len(kv) == 2 {
			headers[strings.TrimSpace(strings.ToLower(kv[0]))] = strings.TrimSpace(kv[1])
		}
	}

	// Step 3: Route handling
	switch {
	case method == "GET" && path == "/":
		fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n\r\n")

	case method == "GET" && strings.HasPrefix(path, "/files/"):
		filename := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(directory, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
			return
		}
		fmt.Fprintf(conn,
			"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
			len(data), data)

	case method == "POST" && strings.HasPrefix(path, "/files/"):
		filename := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(directory, filename)

		// Step 4: Read Content-Length and request body
		contentLength := 0
		if val, ok := headers["content-length"]; ok {
			fmt.Sscanf(val, "%d", &contentLength)
		}

		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
			return
		}

		err = os.WriteFile(filePath, body, 0644)
		if err != nil {
			fmt.Fprint(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
			return
		}

		fmt.Fprint(conn, "HTTP/1.1 201 Created\r\n\r\n")

	default:
		fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}
