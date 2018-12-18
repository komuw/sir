package main

// 1. https://godoc.org/github.com/pa-m/sklearn/cluster
// 2. https://www.cheatography.com/chewxy/cheat-sheets/data-science-in-go-a/

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/komuw/dbscan/proxyd"
	"github.com/pkg/errors"
)

/*
usage:
  go run -race reverse.go localhost:7777 localhost:3333
  curl -vkL localhost:7777
  echo -n "test out the server" | nc localhost 7777
*/

func forward(conn net.Conn) {
	defer conn.Close()

	// TODO: make the buffer growable
	buf := make([]byte, 96)
	reqLen, err := conn.Read(buf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Error reading")
		log.Fatalf("%+v", err)
	}
	_ = reqLen
	log.Println("Reverse read::", buf)
	log.Println("Reverse read2::", string(buf))

	// TODO: since we also want to dbscan the responses, we should make a copy here also.
	client, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		err = errors.Wrap(err, "Reverse Dial failed")
		log.Fatalf("%+v", err)
	}
	defer client.Close()
	err = client.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to set client deadline")
		log.Fatalf("%+v", err)
	}
	log.Printf("Connected to localhost %v\n", conn)

	var myBuf bytes.Buffer
	tee := io.TeeReader(client, &myBuf)

	io.Copy(client, bytes.NewReader(buf))
	io.Copy(conn, tee)

	zsh, _ := ioutil.ReadAll(&myBuf)
	fmt.Println("zsh:::", zsh)
	fmt.Println("zsh2:::", string(zsh))
}

func main() {
	{
		go proxyd.Run()
	}

	if len(os.Args) != 3 {
		log.Fatalf("Reverse Usage %+v listen:port forward:port\n", os.Args[0])
		return
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		err = errors.Wrapf(err, "Reverse failed to setup listener %v", os.Args[1])
		log.Fatalf("%+v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			err = errors.Wrapf(err, "Reverse failed to accept listener %v", os.Args[1])
			log.Fatalf("%+v", err)
		}
		log.Printf("Accepted connection %v\n", conn)
		go forward(conn)
	}
}
