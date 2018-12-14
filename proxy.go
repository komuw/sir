package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
)

/*
usage:
  go run -race proxy.go localhost:7777 localhost:3333
  curl -vkL localhost:7777
  echo -n "test out the server" | nc localhost 7777
*/

func forward(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 100)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Fatal("Error reading:", err)
	}
	_ = reqLen
	log.Println("proxy read::", buf)
	log.Println("proxy read2::", string(buf))

	client, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	defer client.Close()
	log.Printf("Connected to localhost %v\n", conn)

	io.Copy(client, bytes.NewReader(buf))
	io.Copy(conn, client)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage %s listen:port forward:port\n", os.Args[0])
		return
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("ERROR: failed to accept listener: %v", err)
		}
		log.Printf("Accepted connection %v\n", conn)
		go forward(conn)
	}
}
