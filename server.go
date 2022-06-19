package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// main initializes server and starts listening for clients.
func main() {
	addr := flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()

	log.SetFlags(0)
	log.Println("Starting server:", *addr)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleClient(c)
	}
}

// handleClient serves a single client connection.
func handleClient(c net.Conn) {
	defer c.Close()

	log.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		// READ
		msg, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println("Read error: client disconnected,", err)
			return
		}
		log.Printf("From client (%s): %s", c.RemoteAddr().String(), msg)

		// HANDLE
		temp := strings.TrimSpace(string(msg)) // TODO: Implement a function that actually processes the client message and creates a response.
		if temp == "STOP" {
			break
		}

		// WRITE
		c.Write([]byte("HELLO FROM SERVER")) // TODO: respond to client with the response created above.

		// DELAY
		time.Sleep(2 * time.Second) // TODO: change this to a more proper delay (60hz or 30hz or turnbased or whatever).
	}
}
