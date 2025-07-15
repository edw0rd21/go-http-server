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
	"sync/atomic"
)

var directory string

var userCount int64
var activeUsers int64

func main() {
	// Parse --directory flag
	flag.StringVar(&directory, "directory", ".", "Directory to serve files from")
	flag.Parse()

	//Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "4221" // Default for local development
	}

	fmt.Printf("Server running on port %s...\n", port)

	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Printf("Failed to bind to port %s\n", port)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Track user count and active users
		atomic.AddInt64(&userCount, 1)
		atomic.AddInt64(&activeUsers, 1)
		fmt.Printf("New user #%d connected from %s | Active users: %d\n",
			atomic.LoadInt64(&userCount), conn.RemoteAddr(), atomic.LoadInt64(&activeUsers))

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		atomic.AddInt64(&activeUsers, -1)
		fmt.Printf("User disconnected from %s | Active users: %d\n", conn.RemoteAddr(), atomic.LoadInt64(&activeUsers))
	}()

	reader := bufio.NewReader(conn)

	for {
		// Step 1: Read request line
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			// Connection closed or interrupted
			return
		}

		parts := strings.Fields(requestLine)
		if len(parts) < 3 {
			fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
			return
		}
		method := parts[0]
		path := parts[1]

		fmt.Printf("%s request for %s from %s\n", method, path, conn.RemoteAddr())

		// Step 2: Read headers
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
				headers[strings.ToLower(strings.TrimSpace(kv[0]))] = strings.TrimSpace(kv[1])
			}
		}

		// Step 3: Read body (if POST)
		var body []byte
		if method == "POST" {
			if val, ok := headers["content-length"]; ok {
				var contentLength int
				fmt.Sscanf(val, "%d", &contentLength)
				body = make([]byte, contentLength)
				_, err := io.ReadFull(reader, body)
				if err != nil {
					fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
					return
				}
			}
		}

		// Step 4: Handle the request
		switch {
		case method == "GET" && path == "/":
			fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n\r\n")

		case method == "GET" && strings.HasPrefix(path, "/echo/"):
			msg := strings.TrimPrefix(path, "/echo/")
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)

		case method == "GET" && path == "/user-agent":
			ua := headers["user-agent"]
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(ua), ua)

		case method == "GET" && strings.HasPrefix(path, "/files/"):
			filename := strings.TrimPrefix(path, "/files/")
			filePath := filepath.Join(directory, filename)
			data, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				continue
			}
			fmt.Fprintf(conn,
				"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
				len(data), data)

		case method == "POST" && strings.HasPrefix(path, "/files/"):
			filename := strings.TrimPrefix(path, "/files/")
			filePath := filepath.Join(directory, filename)
			err := os.WriteFile(filePath, body, 0644)
			if err != nil {
				fmt.Fprint(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
				return
			}
			fmt.Fprint(conn, "HTTP/1.1 201 Created\r\n\r\n")

		default:
			fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		}

		// (Optional) Respect 'Connection: close'
		if strings.ToLower(headers["connection"]) == "close" {
			return
		}
	}
}
