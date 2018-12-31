package proxyd

import (
	"bytes"
	"log"
	"net"

	"github.com/pkg/errors"
)

func Run(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		err = errors.Wrap(err, "Proxyd Error listening")
		log.Fatalf("\n%+v", err)
	}
	defer l.Close()
	log.Println("Proxyd Listening on " + addr)

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

const nulByte = "\x00"

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
	buf = bytes.Trim(buf, nulByte)
	log.Println("Proxyd read::", buf)
	log.Println("Proxyd read2::", string(buf))

	conn.Write([]byte("Message received."))
}
