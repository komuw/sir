package proxyd

import (
	"log"
	"net"

	"github.com/pkg/errors"
)

/*
usage:
  go run -race proxyd.go
  echo -n "test out the server" | nc localhost 3333
  curl -vkIL localhost:3333
*/

const (
	connHost = "localhost"
	connPort = "3333"
	connType = "tcp"
)

func Run() {
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		err = errors.Wrap(err, "Proxyd Error listening")
		log.Fatalf("\n%+v", err)
	}
	defer l.Close()

	log.Println("Proxyd Listening on " + connHost + ":" + connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			err = errors.Wrap(err, "Proxyd Error accepting")
			log.Fatalf("\n%+v", err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	// TODO: make the buffer growable
	buf := make([]byte, 96)
	reqLen, err := conn.Read(buf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Error reading")
		log.Fatalf("\n%+v", err)
	}
	_ = reqLen
	log.Println("Proxyd read::", buf)
	log.Println("Proxyd read2::", string(buf))

	conn.Write([]byte("Message received."))
}
