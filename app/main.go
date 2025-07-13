package main

import (
	"bufio"
	"flag"
	"fmt"
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

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
		return
	}
	path := parts[1]

	// Read headers, capture User-Agent if needed
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "" {
			break
		}
	}

	switch {
	case path == "/":
		fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n\r\n")
	case strings.HasPrefix(path, "/echo/"):
		body := strings.TrimPrefix(path, "/echo/")
		resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
			len(body), body)
		fmt.Fprint(conn, resp)
	case path == "/user-agent":
		// Not applicable here, but could store user-agent during header loop if needed
		fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n\r\n<user-agent>\r\n")
	case strings.HasPrefix(path, "/files/"):
		filename := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(directory, filename)

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
			return
		}

		resp := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
			len(data), data)
		fmt.Fprint(conn, resp)
	default:
		fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}
