package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer conn.Close()

	// Read and send the client name
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Fprintf(conn, "%s\n", name)

	// Goroutine for receiving messages
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Read and send messages from the user
	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if message != "" {
			fmt.Fprintf(conn, "%s\n", message)
		}
	}
}
