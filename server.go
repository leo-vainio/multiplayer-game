//go:build exclude
// +build exclude

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:1234", "http service address")
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

		go func(c net.Conn) {
			buffer := make([]byte, 1024)

			_, err := c.Read(buffer)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(string(buffer))

			time := time.Now().Format("Monday, 02-Jan-06 15:04:05 MST")
			c.Write([]byte("Hi back!"))
			c.Write([]byte(time))

			c.Close()
		}(c)
	}
}
