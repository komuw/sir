package proxyd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		log.Fatalf("%+v", err)
	}
	defer l.Close()

	log.Println("Proxyd Listening on " + connHost + ":" + connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			err = errors.Wrap(err, "Proxyd Error accepting")
			log.Fatalf("%+v", err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	var connBuf bytes.Buffer
	io.TeeReader(conn, &connBuf)

	connBytes, err := ioutil.ReadAll(&connBuf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to read connBuf")
		log.Fatalf("%+v", err)
	}
	fmt.Println("Reverse connBytes:::", connBytes)
	fmt.Println("Reverse connBytes2:::", string(connBytes))

	conn.Write([]byte("Message received."))
}
